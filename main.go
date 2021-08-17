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
	f, err := os.Open(home + "/data/osm.pbf/shikoku-low.osm.pbf")
	// f, err := os.Open(home + "/data/osm.pbf/japan-low.osm.pbf")
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

	// Create Node list include Relations
	scanner.SkipNodes = true
	scanner.SkipWays = true
	rrelations := 0
	mrway := map[int]string{}
	for scanner.Scan() {
		switch re := scanner.Object().(type) {
		case *osm.Relation:
			if re.Tags.Find("amenity") == "school" {
				rrelations++
				for _, v := range re.Members {
					mrway[int(v.Ref)] = string(v.Type) + "/" + v.Role + "/"
				}
			}
		}
	}
	scanner.Close()

	// Recreate scanner( and rewind file )
	f.Seek(0, 0)
	scanner = osmpbf.New(context.Background(), f, cpu)

	// FeatureCollectionレコード（ヘッダー的なもの）を出力
	file.WriteString("{\"type\":\"FeatureCollection\",\"features\":[\n")

	nodes, ways, relations := 0, 0, 0
	snodes, sways, srelations := 0, 0, 0
	var endl bool
	for scanner.Scan() {

		switch e := scanner.Object().(type) {
		case *osm.Node:
			if e.Tags.Find("amenity") == "school" {
				snodes++
				// 最後のレコード出力時にはカンマを出力しない
				if endl {
					file.WriteString(",\n")
				} else {
					endl = true
				}
				// 要素情報の出力
				file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"Point\",")
				file.WriteString(fmt.Sprintf("\"coordinates\":[%.7f,%.7f]}", e.Lon, e.Lat))
				// 属性文字のエスケープ関連文字の訂正
				if strings.Contains(e.Tags.Find("name"), "\\") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\\", "", -1)))
				} else if strings.Contains(e.Tags.Find("name"), "\n") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\n", "", -1)))
				} else if strings.Contains(e.Tags.Find("name"), "\"") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\"", "　", -1)))
				} else {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", e.Tags.Find("name")))
				}
			}
			nodes++
		case *osm.Way:
			// set way coordinates
			if _, flg := mrway[int(e.ID)]; flg {
				//fmt.Println("Way:", int(e.ID))
				var flon, flat, llon, llat float64
				for i, v := range e.Nodes {
					if i == 0 {
						mrway[int(e.ID)] += fmt.Sprintf("[%.7f,%.7f]", v.Lon, v.Lat)
						flon, flat = v.Lon, v.Lat
					} else {
						mrway[int(e.ID)] += fmt.Sprintf(",[%.7f,%.7f]", v.Lon, v.Lat)
						llon, llat = v.Lon, v.Lat
					}
				}
				if llon == flon && llat == flat {
					mrway[int(e.ID)] += "/close"
				} else {
					mrway[int(e.ID)] += "/open"
				}
			}
			if e.Tags.Find("amenity") == "school" {
				sways++
				file.WriteString(",\n")
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
				if strings.Contains(e.Tags.Find("name"), "\\") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\\", "", -1)))
				} else if strings.Contains(e.Tags.Find("name"), "\n") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\n", "", -1)))
				} else if strings.Contains(e.Tags.Find("name"), "\"") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\"", "　", -1)))
				} else {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", e.Tags.Find("name")))
				}
			}
			ways++
		case *osm.Relation:
			if e.Tags.Find("amenity") == "school" {
				srelations++
				file.WriteString(",\n")

				// 要素情報の出力
				switch e.Tags.Find("type") {
				case "multipolygon":
					file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"MultiPolygon\",")
					file.WriteString("\"coordinates\":[")
				case "multilinestring":
					// not run
					file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"MultiLineString\",")
					file.WriteString("\"coordinates\":[")
				case "polygon":
					// not run
					file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"Polygon\",")
					file.WriteString("\"coordinates\":[")
				case "linestring":
					// not run
					file.WriteString("{\"type\":\"Feature\",\"geometry\":{\"type\":\"LineString\",")
					file.WriteString("\"coordinates\":[")
				}

				i := 0
				lopen := false
				for _, v := range e.Members {
					if v.Type == "way" {
						if way, flg := mrway[int(v.Ref)]; flg {
							// Get way elements
							wayelm := strings.Split(way, "/")
							if i == 0 {
								if wayelm[3] == "open" {
									file.WriteString("[[")
								} else {
									file.WriteString("[")
								}
								i++
							} else {
								if wayelm[3] == "close" && wayelm[1] == "outer" {
									file.WriteString("],[")
								} else {
									file.WriteString(",")
								}
							}
							// file.WriteString(way)
							if wayelm[3] == "close" {
								file.WriteString("[" + wayelm[2] + "]")
							} else {
								file.WriteString(wayelm[2])
								lopen = true
							}
						}
					}
				}
				if lopen {
					file.WriteString("]]]}")
				} else {
					file.WriteString("]]}")
				}

				// 属性文字のエスケープ関連文字の訂正
				if strings.Contains(e.Tags.Find("name"), "\\") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\\", "", -1)))
				} else if strings.Contains(e.Tags.Find("name"), "\n") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\n", "", -1)))
				} else if strings.Contains(e.Tags.Find("name"), "\"") {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\"", "　", -1)))
				} else {
					file.WriteString(fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", e.Tags.Find("name")))
				}

				// For debug
				fmt.Println("Relation Type:", e.Tags.Find("type"))
			}
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

	fmt.Println("snodes:", snodes)
	fmt.Println("sways:", sways)
	fmt.Println("srelations:", srelations)

}
