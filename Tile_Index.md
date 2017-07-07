# Vector Tile Polygon Indexes 

![](https://user-images.githubusercontent.com/10904982/27971958-8f5c2940-6322-11e7-9c95-b66177d5c52d.png)

A vector tile index is essentially just a raycast across the x-axis precreated at a predefined resolution. Currently the resolution I use is a level 9 geohash which is ~20 ft or so. By doing this small simplification and making it a structure along every geohash tile at the highest y coordinate in the tile we can create an easy data structure to hinge are are ray cast structure on and it only requires a geohash to access are ray-cast structure. 

This may seem like it would be a fairly large structure and it is a few hundred kilobytes compressed when stored locally as can be done for a zoom level 10 tile (10s of miles I'd guess), howevery much of our data structure can benefit from the way vector tiles have to be structured. We can also negate the geometry field all together for a properties field rather that holds the positional values of our properties structure. This is beneficial because each value doesn't have to have a header which can really needless add to your tile size for structures like these. (The structure is redudant and doesn't require headers) 

I'll try to spare the details of how the tile is created in the vector_tile protocol but rather focus or list the  performance I get when testing on my computer (not exactly the client but conditions wouldn't be that that different). 

**Its important to remember while this index is stored in a vector tile it differs from vector tiles in the sense that its intended to be used and stored locally on the client read into memory on the fly. Thats why you'll notice most indexs are at higher zooms**

The following example describes one of the worst cases for the tile layers I've created as far as file size goes. It contains 42 different polygon components at a level 10 zoom (33 linear miles diagonal) is approximately ~1mb in vector tile format. The average point in polygon query is almost always less than a micro second. 100,000 point in polygon can be peformed in a < .1 second however your most likely not going to be doing just that if using on this index on the client. 

**Reading into memory / parsing into the data structure of course isn't free but I would still say managable ~10-30 milliseconds to read a tile into memory for use. If using on the client youo shouldn't be having to do reads that often at all, unless right on a tile edge in which case it may be worth while to open both of them.**

### Other potential Use cases
I'm not famiiar with redis at all but I imagine if I get a decent socket implmentation and build the sockets right you could probably point in polygon accross an entire tile set using sockets if desired. It might be pretty expensive to instantiate all the socket instances in memory, but I'm sure this is possible I just have very little experience implementing it.

I've probably said that wrong but you get the idea, sockets holding a tile index can have message (point) directed to them and closed and opened as needed. Something like that. 

Thats about it for now. 
