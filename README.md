# tile_reduce
A tile reduce / batch processing library

#### Goal

While doing geospatial processing in go especially revolving around doing things about tiles I found myself rewriting or poorly implementing the same parts of code over and over again. The nature of the go concurrency model often times makes you take advantage of a piece of data within a go routine which often times leads to sloppy API or process pipelines for performance. In other words go often times makes it hard to break your programs up for different configurations or processes so this module is like the start of an api for multiprocessing jobs about tiles. 

For example two of the processing jobs I've implemented in round about ways are creating tile indexs from one of my other modules for use on the client (hopefully) and for bulk vector tile creation of several zoom layers at once and [fairly quickly if optmized correctly](https://github.com/murphy214/vtile)

So this hopefully will take some of the stuff I've done with tile processing and hopefully get more performance out of this type of pipeline then something written in another language due to its ability to process all the way down by tile after the inialization processing. I've written smaller things like this in python and while easier the top - down nature of the processing and the way in which processes can be hampered by one huge sequential process that isn't using cycles really makes it worthwhile to build something with concurrency in mind. 

