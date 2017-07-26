package tile_surge

import (
	"encoding/json"
	"fmt"
	l "github.com/murphy214/layersplit"
	m "github.com/murphy214/mercantile"
	pc "github.com/murphy214/polyclip"
	"os"
	"strings"
)

// Point represents a point in space.
type Tile_Xmap struct {
	Tile m.TileID
	Xmap map[string][]Yrow
}

type ResponseCoords struct {
	Coords [][][]float64 `json:"coords"`
}

// gets the coordstring into a slice the easiest way I'm aware of
func get_coords_json(stringcoords string) [][][]float64 {
	stringcoords = fmt.Sprintf(`{"coords":%s}`, stringcoords)
	res := ResponseCoords{}
	json.Unmarshal([]byte(stringcoords), &res)

	return res.Coords
}

// Given two slices containg two different geographic layers
// returns 1 slice representing the two layers in which there are split into
// smaller polygons for creating hiearchies e.g. slicing zip codes about counties
func Make_Layer_Xmap(tilemap map[m.TileID][]int, layer []l.Polygon) []Tile_Xmap {
	// creating channel
	c := make(chan Tile_Xmap)

	// iterating through each value in tilemap
	for k, v := range tilemap {
		finds := []l.Polygon{}

		// getting the polygons corresponding to the tileid
		for _, i := range v {
			finds = append(finds, layer[i])
		}

		go func(k m.TileID, finds []l.Polygon, c chan<- Tile_Xmap) {
			c <- Make_Layer_Rect_Xmap(k, finds)
		}(k, finds, c)
	}

	// iterating through each recieved channel output
	count := 0
	total := 0
	progress := 10
	xmaps := []Tile_Xmap{}
	for range tilemap {
		select {
		case msg1 := <-c:
			if count == progress {
				count = 0
				total += progress
				fmt.Printf("[%d/%d]\n", total, len(tilemap))

			}
			xmaps = append(xmaps, msg1)
			count += 1

		}
	}
	return xmaps
}

// Given two slices containg two different geographic layers
// returns 1 slice representing the two layers in which there are split into
// smaller polygons for creating hiearchies e.g. slicing zip codes about counties
func Make_Layer_Xmap_Write(tilemap map[m.TileID][]int, layer []l.Polygon) []Tile_Xmap {
	// creating channel
	c := make(chan string)
	counter := 0
	totalcount := 0
	xmaps := []Tile_Xmap{}

	// iterating through each value in tilemap
	for k, v := range tilemap {
		finds := []l.Polygon{}

		// getting the polygons corresponding to the tileid
		for _, i := range v {
			finds = append(finds, layer[i])
		}

		go func(k m.TileID, finds []l.Polygon, c chan<- string) {
			xtile := Make_Layer_Rect_Xmap(k, finds)
			//fmt.Print("tiles/"+m.Tilestr(xtile.Tile)+".pbf", "\n")
			Make_Vector_Tile_Index(xtile.Xmap, "tiles/"+strings.Replace(m.Tilestr(xtile.Tile), "/", "_", -1)+".pbf")
			c <- string("")
		}(k, finds, c)

		if (counter == 250) || (len(tilemap)-1 == totalcount) {
			// iterating through each recieved channel output
			fmt.Printf("[%d/%d]\n", totalcount, len(tilemap))
			count := 0
			for count < counter {
				select {
				case msg1 := <-c:
					fmt.Print(msg1)
				}
				count += 1
			}
			counter = 0
		}
		totalcount += 1
		counter += 1

	}
	return xmaps
}

// makes layer line that writes out to a file
func Make_Layer_Line_Write(tilemap map[m.TileID][]Line_Edge, fields []string, keysmap map[string]uint32) {
	// iterating through each tileid
	c := make(chan string)
	counter := 0
	totalcount := 0
	for k, v := range tilemap {
		go func(k m.TileID, v []Line_Edge, fields []string, keysmap map[string]uint32, c chan<- string) {
			//fmt.Print("tiles/"+m.Tilestr(xtile.Tile)+".pbf", "\n")
			Make_Vector_Tile_Line_Index(k, v, fields, keysmap)
			c <- string("")
		}(k, v, fields, keysmap, c)

		if (counter == 250) || (len(tilemap)-1 == totalcount) {
			// iterating through each recieved channel output
			fmt.Printf("[%d/%d]\n", totalcount, len(tilemap))
			count := 0
			for count < counter {
				select {
				case msg1 := <-c:
					fmt.Print(msg1)
				}
				count += 1
			}
			counter = 0
		}
		totalcount += 1
		counter += 1

	}
}

func get_polygon(poly pc.Polygon) [][][]float64 {
	newlist3 := [][][]float64{}
	for _, cont := range poly {
		newlist := [][]float64{}
		for _, pt := range cont {
			newlist = append(newlist, []float64{pt.X, pt.Y})
		}
		newlist3 = append(newlist3, newlist)
	}
	return newlist3
}

// given a tileid a set of polygons to that lie within it
// returns the corresponding string a csv file output
func Make_Layer_Rect_Xmap(tile m.TileID, finds []l.Polygon) Tile_Xmap {
	newlist := []l.Polygon{}

	// getting rectangle
	first := l.Polygon{Polygon: Make_Tile_Poly(tile)}
	//first.Polygon.Add(val)
	val := first.Polygon[0]
	val = pc.Contour{val[0], val[1], val[2], val[3], val[2], val[1], val[0]}
	first.Polygon.Add(val)
	// iterating through each found area
	for _, i := range finds {
		i.Polygon.Add(val)
		//if IsReachable(first, i, "INTERSECTION") == true {
		result := first.Polygon.Construct(pc.INTERSECTION, i.Polygon)
		//}

		// adding the the result to newlist if possible
		if len(result) != 0 {
			amap := map[string]string{}
			amap[i.Layer] = i.Area
			amap["tile"] = m.Tilestr(tile)

			//fmt.Print(amap, "\n")
			i.Polygon = result
			i.Layers = amap
			newlist = append(newlist, i)
		} else {
			//	fmt.Print("here\n", first.Polystring, "\n", i.Polystring, "\n")
			//fmt.Print("here\n")
		}

	}
	// linting the output polygons
	stringlist := Lint_Layer_Polygons(newlist)

	// iterating through each value in newlist
	xmaptotal := map[string][]Yrow{}
	for _, i := range stringlist {
		xmap := Make_Xmap_Total(get_coords_json(i[1]), i[0], tile)
		for k, v := range xmap {
			xmaptotal[k] = append(xmaptotal[k], v...)
		}
	}
	valbool := false

	if valbool == true {
		stringlist2 := []string{"LONG,LAT,AREA"}
		for k, v := range xmaptotal {
			x := Get_Middle(k)[0]
			for _, vv := range v {
				area := strings.Replace(vv.Area, ",", "", -1)
				stringlist2 = append(stringlist2, fmt.Sprintf("%f,%f,%s", x, vv.Range[0], area))
				stringlist2 = append(stringlist2, fmt.Sprintf("%f,%f,%s", x, vv.Range[1], area))
			}

		}
		//fmt.Print(xmap, "\n")
		bds := m.Bounds(tile)
		count := 0
		latconst := bds.N

		for count < 100000 {
			count += 1
			pt := RandomPt(bds)
			areat := strings.Replace(Pip_Simple(pt, xmaptotal, latconst), ",", "", -1)
			fmt.Print(areat)
			if areat != "" {
				fmt.Print("Here\n")
				stringlist2 = append(stringlist2, fmt.Sprintf("%f,%f,%s", pt[0], pt[1], areat))
			}
		}

		a := strings.Join(stringlist2, "\n")
		ff, _ := os.Create("d.csv")
		ff.WriteString(a)
		fmt.Print(a, "\n")
	}
	//ff, _ := os.Create("d.csv")
	//ff.WriteString(a)
	return Tile_Xmap{Tile: tile, Xmap: xmaptotal}
}

// lints a polygon layer to add a single layer to the map value
// this makes it so if the area field doesnt' contain a json stirng
// which Make_Layer_Xmap_Write expects it corrects to the correct area syntax
func Lint_Polygon_Layer_Map(layer []l.Polygon) []l.Polygon {
	var newlayer []l.Polygon
	for _, feat := range layer {
		if len(feat.Layers) == 0 {
			area := feat.Area
			layer := feat.Layer
			feat.Area = fmt.Sprintf("{'%s':'%s'}", area, layer)
		} else {
			newlist := []string{}
			for k, v := range feat.Layers {
				newlist = append(newlist, fmt.Sprintf("'%s':'%s'", k, v))
			}
			feat.Area = fmt.Sprintf("{%s}", strings.Join(newlist, ","))
		}
		newlayer = append(newlayer, feat)

	}
	return newlayer
}

// makes tile index from polygons
func Make_Tile_Polygon_Index(config Layer_Config, size int) {
	// making tilemap for polygon
	//layervals := t.Lint_Layer_Polygons(layer)
	newlayer := config.Layer
	firstlayer := config.Layer[0]
	if string(firstlayer.Area[0]) != "{" {
		newlayer = Lint_Polygon_Layer_Map(newlayer)
	}

	tilemap := Make_Tilemap(newlayer, size)

	// creating tiles from a given layer
	Make_Layer_Xmap_Write(tilemap, newlayer)
}

// makes the tile index from lines
func Make_Tile_Line_Index(config Layer_Config, zoom int) {
	// creating tilemap
	tilemap := Make_Tilemap_Lines(config.Line_Layer, zoom)

	// getting fields
	fields := config.Fields
	keysmap := config.Keymap

	// emptying layer config for memory constraints shit
	config = Layer_Config{}

	// making each line out into a tile line
	Make_Layer_Line_Write(tilemap, fields, keysmap)
}
