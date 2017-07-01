package tile_surge

import (
	m "github.com/murphy214/mercantile"
	pc "github.com/murphy214/polyclip"
	"math"
)

// Distance finds the length of the hypotenuse between two points.
// Forumula is the square root of (x2 - x1)^2 + (y2 - y1)^2
func Distance(p1 pc.Point, p2 pc.Point) float64 {
	first := math.Pow(float64(p2.X-p1.X), 2)
	second := math.Pow(float64(p2.Y-p1.Y), 2)
	return math.Sqrt(first + second)
}

func single_point(row pc.Point, bound m.Extrema) []int32 {
	deltax := (bound.E - bound.W)
	deltay := (bound.N - bound.S)

	factorx := (row.X - bound.W) / deltax
	factory := (bound.N - row.Y) / deltay

	xval := int32(factorx * 4096)
	yval := int32(factory * 4096)

	//here1 := uint32((row[0] - bound.w) / (bound.e - bound.w))
	//here2 := uint32((bound.n-row[1])/(bound.n-bound.s)) * 4096
	if xval >= 4095 {
		xval = 4095
	}

	if yval >= 4095 {
		yval = 4095
	}

	return []int32{xval, yval}
}

func Make_Coords(coord []pc.Point, bound m.Extrema) [][]int32 {
	var newlist [][]int32
	//var oldi []float64

	for _, i := range coord {
		newlist = append(newlist, single_point(i, bound))
	}
	return newlist

}

// makes polygon layer for cordinate positions
func Make_Coords_Polygon(polygon pc.Polygon, bound m.Extrema) [][][]int32 {
	var newlist [][][]int32
	//var oldi []float64

	for _, cont := range polygon {
		newcont := [][]int32{}
		for _, i := range cont {
			newcont = append(newcont, single_point(i, bound))
		}
		newlist = append(newlist, newcont)
	}
	return newlist

}
