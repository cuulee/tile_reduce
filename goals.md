# Goals / Approachs to this project.

Tile indexs are designed to be as small as humanly possible when sending to the client and then either a transformation on the vector tile object or a read on the vector tile object will form the desired data structure. Currently for polygons its simple a read on a vector tile index which is fairly quick for a 45 mile diagonal tile. ~1 second.  

However when approaching vector tile indexes for lines we may run into a transformation on the client vs a read every time scenario where it may be worth it to do a one time transformation to make the read quicker simply inflating the data on the vector tile to yield on the information we really want. This takes up more local storage which may be an issue but we will see. Therefore I'll probably protype transformations of lines that would be done on the client in python just get a feel for what were dealing with. 

# Vector Tile Index For Lines. 

Line vector tile indexs are a little more nuanced, we will require more metadata for each feature probably which will mean more fields and also a much larger transformation will be done on the client. Currently I'm thinking something like this for each feature: fields in regular tag set, geometry pointing to a geohash location at the highest precision possible. Then transformation / manipulation on the client. 

## Why geohashs instead of x,y zigzag encoding? 

For one we lose a massive amount of precision zigzag encoding that we can't afford to lose, also a geohash is a probably like a 16 digit string with a precision in the centimeters if not millimeters, as were not rendering and actually using this for an index this is easily the better option. 

### Things that need done

* We will probably need a map structure with gid:{map[string]interface{}} or something like that to reference back from the tilemap of sliced lines. 
* A custom built fuctor / module which won't be to hard. 
