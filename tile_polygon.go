package tile_reduce

import (
	"fmt"
	l "layersplit"
	m "mercantile"
	"os"
	pc "polyclip"
)

// Given two slices containg two different geographic layers
// returns 1 slice representing the two layers in which there are split into
// smaller polygons for creating hiearchies e.g. slicing zip codes about counties
func Make_Layer_Csv(tilemap map[m.TileID][]int, layer []l.Polygon, outfilename string) {
	// creating channel
	c := make(chan string)

	// creating file
	ff, _ := os.Create(outfilename)
	ff.WriteString("LAYERS,COORDS")

	// iterating through each value in tilemap
	for k, v := range tilemap {
		finds := []l.Polygon{}

		// getting the polygons corresponding to the tileid
		for _, i := range v {
			finds = append(finds, layer[i])
		}

		go func(k m.TileID, finds []l.Polygon, c chan<- string) {
			c <- Make_Layer_Rect(k, finds)
		}(k, finds, c)
	}

	// iterating through each recieved channel output
	count := 0
	total := 0
	progress := 10
	for range tilemap {
		select {
		case msg1 := <-c:
			if count == progress {
				count = 0
				total += progress
				fmt.Printf("[%d/%d]\n", total, len(tilemap))

			}
			ff.WriteString("\n" + msg1)
			count += 1

		}
	}

}

// given a tileid a set of polygons to that lie within it
// returns the corresponding string a csv file output
func Make_Layer_Rect(tile m.TileID, finds []l.Polygon) string {
	newlist := []l.Polygon{}

	// getting rectangle
	first := l.Polygon{Polygon: Make_Tile_Poly(tile)}

	// iterating through each found area
	for _, i := range finds {

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

	// creating the final outer square
	val := first.Polygon[0]
	val = pc.Contour{val[0], val[1], val[2], val[3], val[0], val[3], val[2], val[1], val[0]}
	first.Polygon = pc.Polygon{val}
	amap := map[string]string{}
	amap["tile"] = m.Tilestr(tile)
	first.Layers = amap
	if len(newlist) != 0 {
		newlist = append(newlist, first)
	}

	return l.Make_Polygon_String(newlist)
}
