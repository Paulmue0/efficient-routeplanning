# File Format

Data source: OpenStreetMap.
All instances are undirected graphs, and edge cost is one.

* In the first two lines, each contains a single value, where 
  * the first one `n` is the number of nodes
  * the second one `m` is the number of edges

* Then n lines followed and are formatted as `id lat lon`, where
  * `id` is the node id
  * `lat` is the latitude
  * `lon` is the longitude
  
* Then m lines followed and are formatted as `src trg`, where they are node IDs of an edge.

Here is an example:
    
    4
    3
    0 48.667421 9.244557
    1 48.667273 9.244867
    2 48.667598 9.244326
    3 48.667019 9.245514
    0 1
    0 2
    0 3