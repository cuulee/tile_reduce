package tile_reduce

import (
	l "layersplit"
	pc "polyclip"
	"strings"
)

// Overlaps returns whether r1 and r2 have a non-empty intersection.
func Within(big pc.Rectangle, small pc.Rectangle) bool {
	return (big.Min.X <= small.Min.X) && (big.Max.X >= small.Max.X) &&
		(big.Min.Y <= small.Min.Y) && (big.Max.Y >= small.Max.Y)
}

// a check to see if each point of a contour is within the bigger
func WithinAll(big pc.Contour, small pc.Contour) bool {
	totalbool := true
	for _, pt := range small {
		boolval := big.Contains(pt)
		if boolval == false {
			totalbool = false
		}
	}
	return totalbool
}

// creating a list with all of the intersecting contours
// this function returns a list of all the constituent contours as well as
// a list of their keys
func Sweep_Contmap(bb pc.Rectangle, intcont pc.Contour, contmap map[int]pc.Contour) []int {
	newlist := []int{}
	for k, v := range contmap {
		// getting the bounding box
		bbtest := v.BoundingBox()

		// getting within bool
		withinbool := Within(bb, bbtest)

		// logic for if within bool is true
		if withinbool == true {
			withinbool = WithinAll(intcont, v)
		}

		// logic for when we know the contour is within the polygon
		if withinbool == true {
			newlist = append(newlist, k)
		}
	}
	return newlist
}

// getting the outer keys of contours that will be turned into polygons
func make_polygon_list(totalkeys []int, contmap map[int]pc.Contour, relationmap map[int][]int) []pc.Polygon {
	keymap := map[int]string{}
	for _, i := range totalkeys {
		keymap[i] = ""
	}

	// making polygon map
	polygonlist := []pc.Polygon{}
	for k, v := range contmap {
		_, ok := keymap[k]
		if ok == false {
			newpolygon := pc.Polygon{v}
			otherconts := relationmap[k]
			for _, cont := range otherconts {
				newpolygon.Add(contmap[cont])
			}

			// finally adding to list
			polygonlist = append(polygonlist, newpolygon)
		}
	}
	return polygonlist

}

// creates a within map
func Create_Withinmap(contmap map[int]pc.Contour) []pc.Polygon {
	totalkeys := []int{}
	relationmap := map[int][]int{}
	for k, v := range contmap {
		bb := v.BoundingBox()
		keys := Sweep_Contmap(bb, v, contmap)
		relationmap[k] = keys
		totalkeys = append(totalkeys, keys...)
	}

	return make_polygon_list(totalkeys, contmap, relationmap)
}

// lints each polygon
func Lint_Polygons(polygon pc.Polygon) []pc.Polygon {
	// making contour map
	contmap := map[int]pc.Contour{}
	for i, cont := range polygon {
		contmap[i] = cont
	}
	return Create_Withinmap(contmap)

}

// this function takes what would normally be a representitive square or something
// with disorganized contours and sorts them into a layer list
// corresponding to their hole relation
// this function assumes the area field contains the json data
func Lint_Layer_Polygons(layer []l.Polygon) [][]string {
	newlayer := [][]string{}
	for _, i := range layer {
		if len(i.Polygon) > 1 {
			newpolygons := Lint_Polygons(i.Polygon)
			for _, poly := range newpolygons {
				val := strings.Replace(i.Area, ":", "_", 10)
				val = strings.Replace(i.Area, ",", "_", 10)
				newlayer = append(newlayer, []string{val, l.Make_Each_Polygon(poly)})
			}
		} else {
			val := strings.Replace(i.Area, ":", "_", 10)
			val = strings.Replace(i.Area, ",", "_", 10)
			newlayer = append(newlayer, []string{val, l.Make_Each_Polygon(i.Polygon)})
		}
	}
	return newlayer
}
