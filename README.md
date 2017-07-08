# tile_reduce
A tile reduce / batch processing library

[Tile Polygon Index Summary](https://github.com/murphy214/tile_reduce/blob/master/Tile_Index.md)

#### Goal

While doing geospatial processing in go especially revolving around doing things about tiles I found myself rewriting or poorly implementing the same parts of code over and over again. The nature of the go concurrency model often times makes you take advantage of a piece of data within a go routine which often times leads to sloppy API or process pipelines for performance. In other words go often times makes it hard to break your programs up for different configurations or processes so this module is like the start of an api for multiprocessing jobs about tiles. 

For example two of the processing jobs I've implemented in round about ways are creating tile indexs from one of my other modules for use on the client (hopefully) and for bulk vector tile creation of several zoom layers at once and [fairly quickly if optmized correctly](https://github.com/murphy214/vtile)

So this hopefully will take some of the stuff I've done with tile processing and hopefully get more performance out of this type of pipeline then something written in another language due to its ability to process all the way down by tile after the inialization processing. I've written smaller things like this in python and while easier the top - down nature of the processing and the way in which processes can be hampered by one huge sequential process that isn't using cycles really makes it worthwhile to build something with concurrency in mind. 

# File Structure 
#### Misc.
* **line_envelope.go** - Creates the map[xyz][]part_of_line data structure for concurrency about each tile created. 
* **poly_envelope.go** - Creates the map[xyz][]inds_of_layer data structure for concurency about each tile, it should be noted that the values returned are indicies to be referenced from the original layer to prevent unnecessary data duplication. 
* **lint_layer_list.go** - Used when raw polygon rings aren't desired and a structure for the ring ordering is required. (i.e. takes random polygon rings organizes them into simple geometries) 
* **tile_polygon.go** - Sort of the base use case for this repo, outputs a set of polygon geometries sliced about tiles to a csv. This was more of a stepping stone to what the repo was actually built for.

#### Tile Index Part of Codebase
* **tile_xmap.go** - Creates the xmap vector tile structures for a given tile zoom, see documentation on vector tile index for more information.
* **vt_tile_index.go** - Reads / writes a polygon index for a tile to a vector tile representation.
* **xmap.go** - Datastructure used for creating the index, essentially an entire ray-cast for a given tile.

#### Vector Tile Creation Part of Codebase
* **make_tile_layers.go** - Can create how vector tile sets of a given layer (currently only supports one) for both lines and polygons.
* **coords.go** - Used for the raw conversion of float values in map projection to tile coordinates in tile projection (float to int) 
* **geometry.go** - Used for creating the tile geometries for vector tile features for either lines or polygons. 

### The Hard Part 

The hard part isn't getting a pipeline to work with one sequential process thats already been done sort of. The hard part is to make it robust enough to interact with existing APIs ([layersplit](https://github.com/murphy214/layersplit),[geoindex](https://github.com/murphy214/geoindex),and whatever the vector  tile API creation API or structure will look like. I'm not super concerned with vector tile implementations currently because it seems like most renderers have their own set of rules thats tailored to how they produce the tiles.

# Update 

I've basically implemented a decent vector tile API, that renders for mapbox vector tiles in mapbox gl-js but currently the api only supports postgres database pulls and not much configuration is available, however its a pretty decent start, but their are still tons of other things that still need to be implemented. Also, largish, design decisions need to be made, do I want to mill through an entire tileset of several layers at a time (severely complicating every aspect of the API or do i want to create each layer flat into a directory set and combine them about two different directories. This reduces the amount of ram that would be required throughout all the tile processes)

**Also I've never actually implemented a good API in go lol, so this will probably be a bit rocky as far as configuration is concerned, but performance for tile creation for raw lines and polygons is pretty good, I'm not sure how to get tippecannoe to produce actual vector tiles,not raster-like overlays,but for the overlays its about as fast for lines for the entire California roadway dataset.**

Its also worth considering the performance characteristics of different tile layers to be processed for example, sometimes you want sequential processing for huge sets of lines,where you only really process one zoom at a time, however for smaller sets of lines or polygons you may be able to process them all at once in the same go routine. This essentially duplicates the dataset by the number of zooms you have.

### Geometry - Simplification
There also is geometry simplification which won't reduce processing time but will shrink the file size and remove unuseful features at higher zooms. Their are several approachs to attempting this. 


### Gzip - Compression
I can't get mapbox to accept my tile even with the write headers I may try this again later probably just something on my end, its only a few lines of code anyway. 

# Example Polygon
```go
package main

import (
	t "./tile_surge"
)

func main() {
	// getting layer configuration
	c := t.NewConfig("area", "coords", "zip", "polygons", []string{"colorkey"}, []int{2, 3, 4, 5, 6, 7, 8, 9})

	// getting database to be made
	db := t.Make_Layer_DB(c)

	// making the vector tile with the following configuration.
	t.Make_Tile_Layer_Polygon(db)
}

```
#### Output

![](https://user-images.githubusercontent.com/10904982/27981838-c7f3f1aa-6360-11e7-897f-519113ca1dd0.png)

# Example Lines 

```go
package main

import (
	t "./tile_surge"
)

func main() {
	// getting layer configuration
	c := t.NewConfig("gid", "coords", "cal", "lines", []string{"colorkey"}, []int{6, 7, 8, 9, 10, 11, 12})

	// getting database to be made
	db := t.Make_Layer_DB_Line(c)

	// making the vector tile with the following configuration.
	t.Make_Tile_Layer_Line(db)
}
```
#### Output

![](https://user-images.githubusercontent.com/10904982/27981879-df61133a-6361-11e7-87ec-447680163d9f.png)
![](https://user-images.githubusercontent.com/10904982/27981880-df792ff6-6361-11e7-89f8-1a71f912733d.png)
