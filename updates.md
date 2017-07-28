# Important changes the API

I've basically just about settled on using toml data structures for configuration files and essentially dictating everything within this project. 

### Why? 

While the tile indexes don't really have a lot of nuance to them creating vector tiles is a ton of configuration and stuff. I found myself with creaping configuration and struggling to keep it all organized. The toml files I believe, will allow me to have the pretty structure of a configuration file but it can execute completely from said toml file. One can dictate how each layer will be queried and what layers will go in each tile. 

While the toml file currently just sits on top of the api this may soon became something bigger with things like style load outs etc, etc, who knows. 

#### A pragmatic decison to create each layer indepently

I've also decided to probably even when multiple layers are to be put within a tile produce each tile independly and have a function that can run over the directory prefixs and combine them all. **However its worth noting this module is useless cause mapbox doesn't support styling of two layers  in one tile. They need to be there own separate tile.**

#### What a toml file currently looks like.

```
method = "vector"
[datasource]
name = "shit"	
host = "localhost"		
port = 5432				
database = "cal" 	
user = "postgres"			
password = ""			
tablename = "cal"
layertype = "lines"
coordbool = true 	
geomheader = "coords"
tile_type = "vector"

[[layers]]
layer = "level_one"
zooms = [ 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15 ]
fields = [ "gid", "colorkey"]
sql = "select gid,colorkey,coords from cal WHERE highway = 'motorway' OR highway = 'motorway_link' OR highway = 'primary' OR highway = 'primary_link'"


[[layers]]
layer = "level_two"
zooms = [ 8, 9, 10, 11, 12, 13, 14, 15 ]
fields = [ "gid", "colorkey"]
sql = "select gid,colorkey,coords from cal WHERE highway = 'secondary' OR highway = 'secondary_link' OR highway = 'tertiary' OR highway = 'tertiary_link' OR highway = 'trunk' OR highway = 'trunk_link'"


[[layers]]
layer = "level_three"
zooms = [ 10, 11, 12, 13, 14, 15 ]
fields = [ "gid", "colorkey"]
sql = "select gid,colorkey,coords from cal WHERE highway = 'residential' OR highway = 'living_street' OR highway = 'road' OR highway = 'escape' OR highway = 'rest_area'"


[[layers]]
layer = "level_four"
zooms = [ 12, 13, 14, 15 ]
fields = [ "gid", "colorkey"]
sql = "select gid,colorkey,coords from cal WHERE highway = 'disused' OR highway = 'unclassified' OR highway IS NULL"
```

