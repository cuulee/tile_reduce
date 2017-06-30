package tile_reduce

import (
	l "github.com/murphy214/layersplit"
	m "github.com/murphy214/mercantile"
	pc "github.com/murphy214/polyclip"
	"strings"
)

// function for getting the extrema of an alignment
func get_extrema_coords(coords [][]float64) m.Extrema {
	north := -1000.
	south := 1000.
	east := -1000.
	west := 1000.
	lat := 0.
	long := 0.
	for i := range coords {
		lat = coords[i][1]
		long = coords[i][0]

		if lat > north {
			north = lat
		}
		if lat < south {
			south = lat
		}
		if long > east {
			east = long
		}
		if long < west {
			west = long
		}
		//fmt.Print(long, lat, "\n")

	}

	// sorting both lats and longs
	//fmt.Print("e,", east, "w,", west, "s,", south, "n,", north)

	return m.Extrema{S: south, W: west, N: north, E: east}

}

// takes the geohash to  arange
func geoHash2ranges(hash string) (float64, float64, float64, float64) {
	latMin, latMax := -90.0, 90.0
	lngMin, lngMax := -180.0, 180.0
	even := true

	for _, r := range hash {
		// TODO: index step could probably be done better
		i := strings.Index("0123456789bcdefghjkmnpqrstuvwxyz", string(r))
		for j := 0x10; j != 0; j >>= 1 {
			if even {
				mid := (lngMin + lngMax) / 2.0
				if i&j == 0 {
					lngMax = mid
				} else {
					lngMin = mid
				}
			} else {
				mid := (latMin + latMax) / 2.0
				if i&j == 0 {
					latMax = mid
				} else {
					latMin = mid
				}
			}
			even = !even
		}
	}
	if latMin < latMax {
		holder := latMin
		latMax = holder
		latMax = latMin

	}
	if lngMin < lngMax {
		holder := lngMin
		lngMax = holder
		lngMax = lngMin

	}
	return lngMin, lngMax, latMin, latMax
}

// gets the extrema object of  a given geohash
func Geohash_Bounds(ghash string) m.Extrema {
	w, e, s, n := geoHash2ranges(ghash)
	return m.Extrema{S: s, W: w, N: n, E: e}
}

func get_size(tile m.TileID) pc.Point {
	bds := m.Bounds(tile)
	return pc.Point{bds.E - bds.W, bds.N - bds.S}
}

// raw 1d linspace like found in numpy
func linspace(val1 float64, val2 float64, number int) []float64 {
	delta := (val2 - val1) / float64(number)
	currentval := val1
	newlist := []float64{val1}
	for currentval < val2 {
		currentval += delta
		newlist = append(newlist, currentval)
	}

	return newlist
}

func get_middle(tile m.TileID) pc.Point {
	bds := m.Bounds(tile)
	return pc.Point{(bds.E + bds.W) / 2.0, (bds.N + bds.S) / 2.0}
}

func grid_bounds(c2pt pc.Point, c4pt pc.Point, size pc.Point) m.Extrema {
	return m.Extrema{W: c2pt.X - size.X/2.0, N: c2pt.Y + size.Y/2.0, E: c4pt.X + size.X/2.0, S: c4pt.Y - size.Y/2.0}
}

// given a polygon to be tiled envelopes the polygon in corresponding boxes
func envelope_polygon(polygon l.Polygon, size int, intval int, tilemap map[m.TileID][]int) map[m.TileID][]int {
	// getting bds
	bds := fixbounds(polygon)

	// getting all four corners
	c1 := pc.Point{bds.E, bds.N}
	c2 := pc.Point{bds.W, bds.N}
	c3 := pc.Point{bds.W, bds.S}
	c4 := pc.Point{bds.E, bds.S}

	// getting all the tile corners
	c1t := m.Tile(c1.X, c1.Y, size)
	c2t := m.Tile(c2.X, c2.Y, size)
	c3t := m.Tile(c3.X, c3.Y, size)
	c4t := m.Tile(c4.X, c4.Y, size)

	//tilemap := map[m.TileID][]int32{}
	tilemap[c1t] = append(tilemap[c1t], intval)
	tilemap[c2t] = append(tilemap[c2t], intval)
	tilemap[c3t] = append(tilemap[c3t], intval)
	tilemap[c4t] = append(tilemap[c4t], intval)
	sizetile := get_size(c1t)

	//c1pt := get_middle(c1t)
	c2pt := get_middle(c2t)
	//c3pt := get_middle(c3t)
	c4pt := get_middle(c4t)

	gridbds := grid_bounds(c2pt, c4pt, sizetile)
	//fmt.Print(gridbds, sizetile, "\n")
	sizepoly := pc.Point{bds.E - bds.W, bds.N - bds.S}
	xs := []float64{}
	if c2pt.X == c4pt.X {
		xs = []float64{c2pt.X}
	} else {
		xs = []float64{c2pt.X, c4pt.X}

	}
	ys := []float64{}
	if c2pt.Y == c4pt.Y {
		ys = []float64{c2pt.Y}
	} else {
		ys = []float64{c2pt.Y, c4pt.Y}

	}
	if sizetile.X < sizepoly.X {
		number := int((gridbds.E - gridbds.W) / sizetile.X)
		xs = linspace(gridbds.W, gridbds.E, number+1)
	}
	if sizetile.Y < sizepoly.Y {
		number := int((gridbds.N - gridbds.S) / sizetile.Y)
		ys = linspace(gridbds.S, gridbds.N, number+1)
	}

	//totallist := []string{}

	for _, xval := range xs {
		// iterating through each y
		for _, yval := range ys {
			tilemap[m.Tile(xval, yval, size)] = append(tilemap[m.Tile(xval, yval, size)], intval)
		}
	}
	return tilemap

}

// makes the tile polygon
func Make_Tile_Poly(tile m.TileID) pc.Polygon {
	bds := m.Bounds(tile)
	return pc.Polygon{{pc.Point{bds.E, bds.N}, pc.Point{bds.W, bds.N}, pc.Point{bds.W, bds.S}, pc.Point{bds.E, bds.S}}}
}

// fixes bounds somewhere
func fixbounds(polygon l.Polygon) m.Extrema {

	poly := polygon.Polygon
	bds := polygon.Bounds
	for _, i := range poly {
		newbds := i.BoundingBox()
		if newbds.Min.Y < bds.S {
			bds.S = newbds.Min.Y
		}
		if newbds.Max.Y > bds.N {
			bds.N = newbds.Max.Y
		}
		if newbds.Min.X < bds.W {
			bds.W = newbds.Min.X
		}
		if newbds.Max.X > bds.E {
			bds.E = newbds.Max.X
		}
	}
	return bds
}
func Unique(elements []int) []int {
	// Use map to record duplicates as we find them.
	encountered := map[int]bool{}
	result := []int{}
	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func Make_Tilemap(layer []l.Polygon, size int) map[m.TileID][]int {
	tilemap := map[m.TileID][]int{}
	for i, row := range layer {
		tilemap = envelope_polygon(row, size, i, tilemap)
	}

	newtilemap := map[m.TileID][]int{}
	for k, v := range tilemap {
		newtilemap[k] = Unique(v)
	}

	return newtilemap
}
