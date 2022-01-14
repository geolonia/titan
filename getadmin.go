package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

func main() {

	start := time.Now()

	home, _ := os.UserHomeDir()
	// f, err := os.Open(home + "/data/osm.pbf/shikoku-low.osm.pbf")
	f, err := os.Open(home + "/data/osm.pbf/japan-low.osm.pbf")
	// f, err := os.Open(home + "/data/planet-low.osm.pbf")
	if err != nil {
		fmt.Printf("could not open file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	// ファイルを書き込み用にオープン (mode=0666)
	file, err := os.Create("./output.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	cpu := runtime.NumCPU()
	scanner := osmpbf.New(context.Background(), f, cpu)

	// Recreate scanner( and rewind file )
	f.Seek(0, 0)
	scanner = osmpbf.New(context.Background(), f, cpu)

	// FeatureCollectionレコード（ヘッダー的なもの）を出力
	file.WriteString("{\"type\":\"FeatureCollection\",\"features\":[\n")

	nodes, ways, relations := 0, 0, 0
	adminways := 0
	// var endl bool
	for scanner.Scan() {

		switch e := scanner.Object().(type) {
		case *osm.Node:
			nodes++
		case *osm.Way:
			if e.Tags.Find("boundary") == "administrative" && e.Tags.Find("admin_level") == "4" {
				if adminways > 0 {
					file.WriteString(",\n")
				}
				// 要素情報の出力
				if e.Polygon() {
					file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"Polygon\",")
					file.WriteString("\"coordinates\":[[")
				} else {
					file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"LineString\",")
					file.WriteString("\"coordinates\":[")
				}

				for i, v := range e.Nodes {
					if i > 0 {
						file.WriteString(",")
					}
					file.WriteString(fmt.Sprintf("[%.7f,%.7f]", v.Lon, v.Lat))
				}
				if e.Polygon() {
					file.WriteString("]]}")
				} else {
					file.WriteString("]}")
				}

				// file.WriteString(fmt.Sprintf("\"coordinates\":[%.7f,%.7f]}", e.Lon, e.Lat))
				// 属性文字のエスケープ関連文字の訂正
				if strings.Contains(e.Tags.Find("admin_level"), "\\") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"admin_level\":\"%s\"}}", strings.Replace(e.Tags.Find("admin_level"), "\\", "", -1)))
				} else if strings.Contains(e.Tags.Find("admin_level"), "\n") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"admin_level\":\"%s\"}}", strings.Replace(e.Tags.Find("admin_level"), "\n", "", -1)))
				} else if strings.Contains(e.Tags.Find("admin_level"), "\"") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"admin_level\":\"%s\"}}", strings.Replace(e.Tags.Find("admin_level"), "\"", "　", -1)))
				} else {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"admin_level\":\"%s\"}}", e.Tags.Find("admin_level")))
				}
				adminways++
			}
			ways++
		case *osm.Relation:
			relations++
		}

	}
	// FeatureCollection終端を出力
	file.WriteString("]}\n")

	if err := scanner.Err(); err != nil {
		fmt.Printf("scanner returned error: %v", err)
		os.Exit(1)
	}

	end := time.Now()

	fmt.Println("Start:", start)
	fmt.Println("End  :", end)
	fmt.Println("Elapsed:", end.Sub(start))

	fmt.Println("nodes:", nodes)
	fmt.Println("ways:", ways)
	fmt.Println("relations:", relations)

	fmt.Println("adminways:", adminways)

}
