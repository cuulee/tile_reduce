# tile_reduce
A tile reduce / batch processing library

#### Goal

While doing geospatial processing in go especially revolving around doing things about tiles I found myself rewriting or poorly implementing the same parts of code over and over again. The nature of the go concurrency model often times makes you take advantage of a piece of data within a go routine which often times leads to sloppy API or process pipelines for performance. In other words go often times makes it hard to break your programs up for different configurations or processes so this module is like the start of an api for multiprocessing jobs about tiles. 

For example two of the processing jobs I've implemented in round about ways are creating tile indexs from one of my other modules for use on the client (hopefully) and for bulk vector tile creation of several zoom layers at once and [fairly quickly if optmized correctly](https://github.com/murphy214/vtile)

So this hopefully will take some of the stuff I've done with tile processing and hopefully get more performance out of this type of pipeline then something written in another language due to its ability to process all the way down by tile after the inialization processing. I've written smaller things like this in python and while easier the top - down nature of the processing and the way in which processes can be hampered by one huge sequential process that isn't using cycles really makes it worthwhile to build something with concurrency in mind. 

### The Hard Part 

The hard part isn't getting a pipeline to work with one sequential process thats already been done sort of. The hard part is to make it robust enough to interact with existing APIs ([layersplit](https://github.com/murphy214/layersplit),[geoindex](https://github.com/murphy214/geoindex),and whatever the vector  tile API creation API or structure will look like. I'm not super concerned with vector tile implementations currently because it seems like most renderers have their own set of rules thats tailored to how they produce the tiles.

# Update 

I've basically implemented a decent vector tile API, that renders for mapbox vector tiles in mapbox gl-js but currently the api only supports postgres database pulls and not much configuration is available, however its a pretty decent start, but their are still tons of other things that still need to be implemented. Also, largish, design decisions need to be made, do I want to mill through an entire tileset of several layers at a time (severely complicating every aspect of the API or do i want to create each layer flat into a directory set and combine them about two different directories. This reduces the amount of ram that would be required throughout all the tile processes)

Its also worth considering the performance characteristics of different tile layers to be processed for example, sometimes you want sequential processing for huge sets of lines,where you only really process one zoom at a time, however for smaller sets of lines or polygons you may be able to process them all at once in the same go routine. This essentially duplicates the dataset by the number of zooms you have.

### Geometry - Simplification
There also is geometry simplification which won't reduce processing time but will shrink the file size and remove unuseful features at higher zooms. Their are several approachs to attempting this. 


### Gzip - Compression
I can't get mapbox to accept my tile even with the write headers I may try this again later probably just something on my end, its only a few lines of code anyway. 

