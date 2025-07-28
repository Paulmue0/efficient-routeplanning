package pathfinding

import (
	"fmt"
	"slices"

	collection "github.com/PaulMue0/efficient-routeplanning/Collection"
	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

type ContractionHierarchies struct {
	ContractionOrder []graph.VertexId
	EdgeDifferences  *collection.PriorityQueue[graph.VertexId]
}

func NewContractionHierarchies() *ContractionHierarchies {
	co := make([]graph.VertexId, 0)
	ed := collection.NewPriorityQueue[graph.VertexId]()

	return &ContractionHierarchies{co, ed}
}

// Preprocessing Steps:
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

func Shortcuts(g *graph.Graph, v graph.VertexId, insertFlag bool) int {
	shortcutsFound := 0
	neighbors, _ := g.Neighbors(v)

	for i := 0; i < len(neighbors)-1; i++ {
		for j := i + 1; j < len(neighbors); j++ {
			path, cost, _ := DijkstraShortestPath(g, neighbors[i].Id, neighbors[j].Id)
			if slices.Contains(path, v) && len(path) == 3 {
				if insertFlag {
					g.AddEdge(neighbors[i].Id, neighbors[j].Id, int(cost), true, v)
				}
				shortcutsFound++
			}
		}
	}
	return shortcutsFound
}

// ED(v) is the number of shortcuts that would need to be added if contracting v
// minus the number of edges that would get contracted (degree of v).
func EdgeDifference(g *graph.Graph, v graph.VertexId) int {
	degree, _ := g.Degree(v)
	shortcuts := Shortcuts(g, v, false)

	fmt.Println(degree, shortcuts, shortcuts-degree)
	return shortcuts - degree
}

func Contract(g *graph.Graph, v graph.VertexId, contractionOrder []graph.VertexId) {
	// update priority
	//:would
	// get highest prio note
	//
	// contract this node
	// 	->
	// 	-> insert shortcuts for this node
	//
}
