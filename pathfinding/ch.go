package pathfinding

import (
	"fmt"
	"slices"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

// get shortcuts...
//
// 1. needs (graph, node)
// should find all shortcuts that would occur when note is contracted
//
// 2. get neighbors of node
// 3. check shortest path between all neighbors
// 	- if the shortest path contains the original node -> shortcuts added +1
// return num of shortcuts and also the node tuples of the two shortcuts
//
//

func NumShortcuts(g *graph.Graph, v graph.VertexId) int {
	shortcutsFound := 0
	neighbors, _ := g.Neighbors(v)

	for i := 0; i < len(neighbors)-1; i++ {
		for j := 1; j < len(neighbors); j++ {
			path, _ := DijkstraShortestPath(g, neighbors[i].Id, neighbors[j].Id)
			if slices.Contains(path, v) {
				fmt.Println(i, j, path)
				shortcutsFound++
			}
		}
	}
	return shortcutsFound
}
