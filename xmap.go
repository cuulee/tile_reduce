package tile_reduce

// This file creates the most important data stucture within this alg.
// The xmap data stucture that pairs x intercepts within a map so that
// they can be utilized in an easy point in polygon check against a points x value
// Essentially this creates the ray casting stucture that can check points
// much much faster then a simple iterate through the coordinates and check

// See the coordinate structure below for a more illustrative definiation
//    123
//0   -
//1	   -
//2    -
//3	  -
//4     -
//5    -
//6    -
//7     -
// Output data structure: {1:[[0,3]],
//						   2:[[1,2],[5,6]],
//						   3:[[4,7]]}

// Except your keys would be geohashs and your map would contain a latitude const value to use.
// This stucture allows you to ray cast easily at a defined resolution on an interminenate amount of points

import (
	//"fmt"
	//"fmt"
	geo "github.com/paulmach/go.geo"
	"math"
	"math/rand"
	m "mercantile"
	"sort"
	"strings"
)

// Point represents a point in space.
type Size struct {
	deltaX float64
	deltaY float64
	linear float64
}

// Point represents a point in space.
type Extrema struct {
	s float64
	w float64
	n float64
	e float64
}

// Point represents a point in space.
type Yrow struct {
	Range []float64
	Area  string
	Y     float64
}

// Point represents a point in space.
type Point struct {
	X float64
	Y float64
}

// gets the slope of two points along a line
// if statement logic accounts for undefined corner case
func get_slope(pt1 Point, pt2 Point) float64 {
	if pt1.X == pt2.X {
		return 1000000.0
	}
	return (pt2.Y - pt1.Y) / (pt2.X - pt1.X)
}

// iteroplates the position of y based on x of the location between two points
// this function accepts m the slope to keep it from recalculating
// what could be several hundred/thousand times between two points
func interp(pt1 Point, pt2 Point, m float64, x float64) Point {
	y := (x-pt1.X)*m + pt1.Y
	return Point{x, y}
}

// linting an alignment to make sure start matches end
func lint_coords(coords [][]float64) [][]float64 {
	if coords[0][0] != coords[len(coords)-1][0] {
		//fmt.Print(coords[0][0])
		coords = append(coords, coords[0])
	}
	return coords
}

// returns a data structure contain the size of your deltaX deltaY and linear distance (accross the box)
// given a geohash return the linear distance from one corner to another
func get_size_geohash(ghash string) Size {
	ex := get_extrema_ghash(ghash)
	size := math.Sqrt(math.Pow(ex.n-ex.s, 2) + math.Pow(ex.e-ex.w, 2))
	return Size{ex.e - ex.w, ex.n - ex.s, size}
}

// function for getting the extrema of an alignment
func get_extrema_coords2(coords [][]float64, sizes Size) Extrema {
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
	return Extrema{south - sizes.deltaY*4, west - sizes.deltaX*4, north + sizes.deltaY*4, east + sizes.deltaX*4}

}

func geoHash2ranges2(hash string) (float64, float64, float64, float64) {
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

	return lngMin, lngMax, latMin, latMax
}

// gets the extrema object of  a given geohash
func get_extrema_ghash(ghash string) Extrema {
	w, e, s, n := geoHash2ranges2(ghash)
	return Extrema{s, w, n, e}
}

// gets the extrema object of  a given geohash
func Get_Middle(ghash string) []float64 {
	w, e, s, n := geoHash2ranges2(ghash)
	return []float64{(w + e) / 2.0, (s + n) / 2.0}
}

// fills in the x values from a point on the polygon line representation to
// the next point on the line and fills the appopriate geohashs between them
// this section can be though of as the raycasting solver for geohash level we desire
// currently defaults to size 9
func fill_x_values(pt1 Point, pt2 Point, sizes Size, latconst float64) map[string]float64 {
	// creating temporary map
	tempmap := map[string]float64{}

	// getting the geohashs for the relevant points
	ghash1 := geo.NewPoint(pt1.X, pt1.Y).GeoHash(9)
	ghash2 := geo.NewPoint(pt2.X, pt2.Y).GeoHash(9)

	// decoding each geohash to a point in space
	long1 := geo.NewPointFromGeoHash(ghash1).Lng()
	long2 := geo.NewPointFromGeoHash(ghash2).Lng()

	// sorting are actual longs
	longs := []float64{pt1.X, pt2.X}
	sort.Float64s(longs)

	// getting potential longs
	// this is mainly just to check the first one
	potlongs := []float64{long1, long2}
	sort.Float64s(potlongs)
	xcurrent := potlongs[0]

	m := get_slope(pt1, pt2)
	pt := Point{0.0, 0.0}
	var ghash string
	//total := [][]float64{}
	for xcurrent < longs[1] {
		pt = interp(pt1, pt2, m, xcurrent)
		ghash = geo.NewPoint(xcurrent, latconst).GeoHash(9)

		if (xcurrent > longs[0]) && (xcurrent < longs[1]) {
			//total = append(total, []float64{xcurrent, pt.Y})
			tempmap[ghash] = pt.Y
		}
		xcurrent += sizes.deltaX
	}
	return tempmap
}

// given a set of x,y coordinates returns a map string that will be used as the base
// string for constructing our geohash tables
// this is essentially the most important data structure for the algorithm
func Make_Xmap(coords [][]float64, areastring string, bds m.Extrema) map[string][]Yrow {
	// quick lint
	N := bds.N

	// sizing a single geohash like this for now
	ghash := geo.NewPoint(coords[0][0], coords[0][1]).GeoHash(9)

	sizes := get_size_geohash(ghash)

	// linting coord values
	//coords = lint_coords(coords)
	coords = append(coords, coords[0])

	// getting coords extrema

	// intialization variables
	latconst := N
	pt := Point{0.0, 0.0}
	oldpt := Point{0.0, 0.0}
	count := 0
	row := []float64{}
	topmap := map[string][]float64{}
	topmap["latconst"] = append(topmap["float64"], latconst)

	// iterating through each coordinate collecting each fill_x_values output
	for i := range coords {
		row = coords[i]
		pt = Point{row[0], row[1]}
		if count == 0 {
			count = 1
		} else {
			//go func(oldpt Point,pt Point,sizes Size,latconst float64,ccc chan<- []float64) {
			tempmap := fill_x_values(oldpt, pt, sizes, latconst)
			for k, v := range tempmap {
				topmap[k] = append(topmap[k], v)

			}

		}

		oldpt = pt

	}

	// creating outer level map but sorting
	newmap := map[string][]Yrow{}
	for k, v := range topmap {
		sort.Float64s(v)
		newlist := []int{}
		//newlist2 := [][]float64{}
		yrows := []Yrow{}
		for i := range v {
			newlist = append(newlist, i)
			if len(newlist) == 2 {
				//newlist2 = append(newlist2, []float64{v[newlist[0]], v[newlist[1]]})
				yrows = append(yrows, Yrow{Range: []float64{v[newlist[0]], v[newlist[1]]}, Area: areastring, Y: v[newlist[0]]})
				newlist = []int{}
			}
		}
		newmap[k] = yrows
		//fmt.Print(newlist2, "\n")
		//topmap[k] = v

	}
	//fmt.Print(len(topmap), "\n\n")
	return newmap

}

// fixes the holes
func fix_holes(xmaptotal map[string][]Yrow) map[string][]Yrow {
	newlist := [][]int{}
	for k, v := range xmaptotal {
		for ii, testhole := range v {
			boolval := false
			pos := 0
			for i, vnext := range v {
				if (vnext.Range[0] < testhole.Y) && (vnext.Range[1] > testhole.Y) {
					pos = i
					boolval = true
				}
			}
			if boolval == true {
				newlist = append(newlist, []int{ii, pos})
			}
		}

		for _, row := range newlist {
			v[row[0]].Range[1] = v[row[1]].Range[0]
			v[row[1]].Range[0] = v[row[0]].Range[1]
		}
		xmaptotal[k] = v

	}
	return xmaptotal
}

func Make_Xmap_Total(coords [][][]float64, area string, tile m.TileID) map[string][]Yrow {
	// getting north value
	bds := m.Bounds(tile)

	// fixing area
	area = area[1 : len(area)-1]
	vals := strings.Split(area, ",")

	mymap := map[string]string{}
	keys := []string{}
	for _, row := range vals {
		kv := strings.Split(row, ":")
		mymap[kv[0]] = kv[1]
		keys = append(keys, kv[0])
	}
	sort.Strings(keys)

	newlist := []string{}
	for _, i := range keys {
		newlist = append(newlist, i+":"+mymap[i])
	}
	area = "{" + strings.Join(newlist, ",") + "}"

	// combining all xmaps
	xmaptotal := map[string][]Yrow{}
	for _, coord := range coords {
		xmaptemp := Make_Xmap(coord, area, bds)
		for k, v := range xmaptemp {
			xmaptotal[k] = append(xmaptotal[k], v...)
		}
	}

	// fixing holes
	xmaptotal = fix_holes(xmaptotal)

	return xmaptotal
}

func Pip_Simple(pt []float64, topmap map[string][]Yrow, latconst float64) string {
	ycollisions := topmap[geo.NewPoint(pt[0], latconst).GeoHash(9)]
	for _, i := range ycollisions {

		if (i.Range[0] <= pt[1]) && (i.Range[1] >= pt[1]) {
			return i.Area
		}

	}
	return ""
}

func RandomPt(bds m.Extrema) []float64 {
	deltax := math.Abs(bds.W - bds.E)
	deltay := math.Abs(bds.N - bds.S)
	return []float64{(rand.Float64() * deltax) + bds.W, (rand.Float64() * deltay) + bds.S}
}
