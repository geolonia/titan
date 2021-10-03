package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

// define file
// const pbfname string = "/data/osm.pbf/shikoku-low.osm.pbf"
const pbfname string = "/data/osm.pbf/japan-low.osm.pbf"

// const pbfname string = "/data/osm.pbf/planet-low.osm.pbf"
// const pbfname string = "/data/osm.pbf/hokkaido-low.osm.pbf"
const outfname string = "./output.json"
const debugfname string = "./debug.json"
const debugfname1 string = "./debug1.json"

// define Tag
const tagname string = "amenity"
const tagval string = "school"

// ====================================
// マルチポリゴンのみ出力するバージョン
// 2021.9.21 K.Takamoto
// ====================================
func main() {

	// get start time
	start := time.Now()

	// ====================================
	// 0.create debug file
	// ====================================
	dfile, err := os.Create(debugfname)
	if err != nil {
		fmt.Printf("could not create debug file: %v", err)
		os.Exit(1)
	}
	defer dfile.Close()

	dfile1, err := os.Create(debugfname1)
	if err != nil {
		fmt.Printf("could not create debug file1: %v", err)
		os.Exit(1)
	}
	defer dfile1.Close()

	// ====================================
	// 1.open osm.pbf file and create scanner
	// ====================================
	home, _ := os.UserHomeDir()
	f, err := os.Open(home + pbfname)
	if err != nil {
		fmt.Printf("could not open file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	// get number of cpu and create scanner
	cpu := runtime.NumCPU()
	scanner := osmpbf.New(context.Background(), f, cpu)

	// ====================================
	// 2.create Node list include Relations(mrway)
	// ====================================
	scanner.SkipNodes = true
	scanner.SkipWays = true
	mrway := map[int]string{}
	// for debug
	mdebug := map[string]int{}
	for scanner.Scan() {
		switch e := scanner.Object().(type) {
		case *osm.Relation:
			if e.Tags.Find(tagname) == tagval {
				for _, v := range e.Members {
					k := e.Tags.Find("type") + "/" + string(v.Type) + "/" + v.Role + "/"
					mrway[int(v.Ref)] = k
					// for debug
					mdebug[k] += 1
				}
			}
		}
	}
	scanner.Close()

	// ====================================
	// 3.add coordinate and set open/close to mrway
	// ====================================
	f.Seek(0, 0)
	scanner = osmpbf.New(context.Background(), f, cpu)
	scanner.SkipNodes = true
	scanner.SkipRelations = true
	for scanner.Scan() {
		switch re := scanner.Object().(type) {
		case *osm.Way:
			// set way coordinates
			if _, flg := mrway[int(re.ID)]; flg {
				//fmt.Println("Way:", int(e.ID))
				var flon, flat, llon, llat float64
				for i, v := range re.Nodes {
					if i == 0 {
						mrway[int(re.ID)] += fmt.Sprintf("[%.7f,%.7f]", v.Lon, v.Lat)
						flon, flat = v.Lon, v.Lat
					} else {
						mrway[int(re.ID)] += fmt.Sprintf(",[%.7f,%.7f]", v.Lon, v.Lat)
						llon, llat = v.Lon, v.Lat
					}
				}
				// start and last coordinate
				mrway[int(re.ID)] += fmt.Sprintf("/[%.7f,%.7f]", flon, flat)
				mrway[int(re.ID)] += fmt.Sprintf("/[%.7f,%.7f]", llon, llat)
				//
				if llon == flon && llat == flat {
					mrway[int(re.ID)] += "/close"
				} else {
					mrway[int(re.ID)] += "/open"
				}
			}
		}
	}
	scanner.Close()

	// ====================================
	// 4.create GeoJSON file
	// ====================================
	file, err := os.Create(outfname)
	if err != nil {
		fmt.Printf("could not create file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	// for debug
	nodes, ways, relations := 0, 0, 0
	snodes, sways, srelations := 0, 0, 0
	var endl bool

	// ====================================
	// 5.Recreate scanner and write GeoJSON
	// ====================================
	f.Seek(0, 0)
	scanner = osmpbf.New(context.Background(), f, cpu)

	// For debug
	// scanner.SkipNodes = true
	// scanner.SkipWays = true
	// scanner.SkipRelations = true

	// write "FeatureCollection" record
	file.WriteString("{\"type\":\"FeatureCollection\",\"features\":[\n")

	for scanner.Scan() {
		switch e := scanner.Object().(type) {
		case *osm.Node:
			if e.Tags.Find(tagname) == tagval {
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
			if e.Tags.Find(tagname) == tagval {
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
			// target Tag
			if e.Tags.Find(tagname) == tagval {
				// GeoJSON output string
				var geojson string

				srelations++
				geojson += ",\n"
				// ここはTypeで分岐
				if e.Tags.Find("type") == "multipolygon" {
					// MultiPolygon
					geojson += "{\"type\":\"Feature\",\"geometry\":{\"type\":\"MultiPolygon\","
				} else {
					// site(MultiLineString)
					geojson += "{\"type\":\"Feature\",\"geometry\":{\"type\":\"MultiLineString\","
				}
				geojson += "\"coordinates\":["

				// for debug
				dfile.WriteString("Element:" + strconv.Itoa(int(e.ID)) + "/" + e.Tags.Find("name") + "\n")
				dfile1.WriteString("Element:" + strconv.Itoa(int(e.ID)) + "/" + e.Tags.Find("name") + "\n")

				// For debug
				if e.Tags.Find("name") == "札幌市立東橋小学校" {
					relations++
				}

				var cntcoord int
				//var startcoord string
				var boutfile, bouter, binner, bsite, bopen bool
				// ******************************
				// start coordinates
				// ******************************
				startchain := map[string]string{} // start coordinate chain(dictionary)
				endchain := map[string]string{}   // end coordinate chain(dictionary)
				openel := map[string]string{}     // open element dictionary
				var innerel []string              // inner element

				for _, v := range e.Members {
					// element kind "Way" only processing
					if v.Type == "way" {
						// set output flag
						boutfile = true
						// get "Way" information from mrway( dictionary )
						if way, flg := mrway[int(v.Ref)]; flg {

							// for debug
							dfile.WriteString(way + "\n")

							// split way information
							wayelm := strings.Split(way, "/")
							// mrway delimter "/".
							// wayelm array format.
							//  [0]: relation type("multipolygon" ,"site")
							//  [1]: element type( "way" ,"node" )
							//  [2]: role( "outer","inner","entrance","perimeter","label","" )
							//  [3]: coordinates
							//  [4]: start coordinate
							//  [5]: end coordinate
							//  [6]: open/close area( "open"/"close" )

							//
							if wayelm[0] == "multipolygon" {
								// for debug
								dfile1.WriteString(wayelm[2] + "/" + wayelm[6] + "/S:" + wayelm[4] + "/E:" + wayelm[5] + "\n")
								// MultiPolygon
								if wayelm[6] == "close" {
									// closed element
									if wayelm[2] == "outer" || wayelm[2] == "" {
										// =======================
										// closed outer element
										// =======================
										if bouter {
											geojson += "]"
										}
										if cntcoord > 0 {
											geojson += ","
										}
										geojson += "[[" + wayelm[3] + "]"
										if binner {
											geojson += "]"
										}
										for _, inner := range innerel {
											geojson += ",[" + inner + "]"
										}
										bouter = true
										binner = false
									} else {
										// =======================
										// closed inner element
										// =======================
										if !bouter {
											innerel = append(innerel, wayelm[3])
											continue
										}
										if cntcoord > 0 {
											geojson += ","
										}
										geojson += "[" + wayelm[3] + "]"
										bouter = false
										binner = true
									}
									cntcoord++
								} else {
									// *******************************
									// open element
									// *******************************
									_, estart := startchain[wayelm[4]]
									_, eend := endchain[wayelm[5]]
									if estart || eend {
										// =============================
										// coordinate exists start or end chain.
										// Reverse coordinates and strat/end chain.
										// =============================
										// add element to chain( key:start coordinate , value:last coordinate)
										startchain[wayelm[5]] = wayelm[4]
										endchain[wayelm[4]] = wayelm[5]
										// add element( key:start coordinate , value:coordinates)
										var revcoo string
										coo := strings.Split(wayelm[3], ",[")
										for j := len(coo); j > 0; j-- {
											if coo[j-1][0] != '[' {
												revcoo += "["
											}
											revcoo += coo[j-1]
											if j > 1 {
												revcoo += ","
											}
										}
										openel[wayelm[5]] = revcoo
									} else {
										// add element to chain( key:start coordinate , value:last coordinate)
										startchain[wayelm[4]] = wayelm[5]
										endchain[wayelm[5]] = wayelm[4]
										// add element( key:start coordinate , value:coordinates)
										openel[wayelm[4]] = wayelm[3]
									}
									bopen = true
								}
							} else {
								// site(MultiLineString)
								if cntcoord > 0 {
									geojson += ","
								}
								geojson += "[" + wayelm[3] + "]"
								cntcoord++
								bsite = true
							}
						}
					} else {
						break
					}
				}
				// open element
				if bopen {
					// seek 'startchain'
					var coordinate string
					for key, val := range startchain {
						// set start coordinates
						coordinate = openel[key]
						delete(startchain, key)
						delete(openel, key)
						// set length
						chainlen := len(startchain)
						for j := 0; j < chainlen; j++ {
							if _, exi := openel[val]; exi {
								coordinate += "," + openel[val]
								delete(openel, val)
								oval := val
								val = startchain[val]
								delete(startchain, oval)
							} else {
								// failure search chain
								break
							}
						}
						if cntcoord > 0 {
							geojson += "],"
						}
						geojson += "["
						geojson += "[" + coordinate + "]"
						if len(startchain) > 0 {
							geojson += "],"
						}
					}
					cntcoord++
				}
				// ******************************
				// close coordinates
				// ******************************
				if !bsite {
					geojson += "]"
				}
				geojson += "]}"

				// 属性文字のエスケープ関連文字の訂正
				if strings.Contains(e.Tags.Find("name"), "\\") {
					geojson += fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\\", "", -1))
				} else if strings.Contains(e.Tags.Find("name"), "\n") {
					geojson += fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\n", "", -1))
				} else if strings.Contains(e.Tags.Find("name"), "\"") {
					geojson += fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", strings.Replace(e.Tags.Find("name"), "\"", "　", -1))
				} else {
					geojson += fmt.Sprintf(",\"properties\":{\"name\":\"%s\"}}", e.Tags.Find("name"))
				}

				// write file
				if boutfile {
					file.WriteString(geojson)
				}

				fmt.Println("Relation Type:", e.Tags.Find("type"))
			}
			relations++
		}

	}
	scanner.Close()
	// FeatureCollection終端を出力
	file.WriteString("]}\n")

	// result
	end := time.Now()
	fmt.Println("Start:", start, "\nEnd:", end, "\nElapsed:", end.Sub(start))
	fmt.Println("nodes[", nodes, "] ways[", ways, "] relations[", relations, "]\nsnodes[", snodes, "] sways[", sways, "] srelations[", srelations, "]")
}
