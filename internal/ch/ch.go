package ch

import (
	"container/heap"
	"fmt"
	"slices"

	pathfinding "github.com/PaulMue0/efficient-routeplanning/internal/pathfinding"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
	collection "github.com/PaulMue0/efficient-routeplanning/pkg/collection/heap_gen"
)

type ContractionHierarchies struct {
	NumShortcutsAdded int
	ContractionOrder  []graph.VertexId
	Priorities        *collection.PriorityQueue[graph.VertexId]
	UpwardsGraph      *graph.Graph
	DownwardsGraph    *graph.Graph
}

var ShortcutsAdded = 0

func NewContractionHierarchies() *ContractionHierarchies {
	co := make([]graph.VertexId, 0)
	ed := collection.NewPriorityQueue[graph.VertexId]()
	ug := graph.NewGraph()
	dg := graph.NewGraph()

	return &ContractionHierarchies{0, co, ed, ug, dg}
}

func (c *ContractionHierarchies) Preprocess(g *graph.Graph) {
	c.InitializePriority(g)
	// WHILE G HAS NODES:
	for len(g.Vertices) >= 1 {
		item := heap.Pop(c.Priorities).(*collection.Item[graph.VertexId])
		v := c.Priorities.GetValue(item)

		neighbors, _ := g.Neighbors(v)
		c.Contract(g, v)

		// Update ONLY the priorities of the neighbors
		for _, neighbor := range neighbors {
			if _, ok := g.Vertices[neighbor.Id]; ok { // Check if neighbor hasn't been contracted yet
				newNeighborPrio := Priority(g, neighbor.Id)
				c.Priorities.UpdatePriority(neighbor.Id, newNeighborPrio)
			}
		}
	}
}

func (c *ContractionHierarchies) InitializePriority(g *graph.Graph) {
	for _, v := range g.Vertices {
		c.Priorities.PushWithPriority(v.Id, Priority(g, v.Id))
	}
}

func (c *ContractionHierarchies) UpdatePriorities(g *graph.Graph) {
	// TODO: for all nodes in parallel run update edge differences
	for _, v := range g.Vertices {
		c.Priorities.UpdatePriority(v.Id, Priority(g, v.Id))
	}
}

func (c *ContractionHierarchies) Contract(g *graph.Graph, v graph.VertexId) {
	Shortcuts(g, v, true)
	// TODO: Needs wait group when run in parallel!
	c.ContractionOrder = append(c.ContractionOrder, v)
	c.InsertInUpwardsOrDownwardsGraph(g, v)

	if err := g.RemoveVertex(v); err != nil {
		panic(fmt.Sprintf("critical error removing vertex %v: %v.\n Edges: %v,\n Vertices %v", v, err, g.Edges[v], g.Vertices))
	}
}

func (c *ContractionHierarchies) InsertInUpwardsOrDownwardsGraph(g *graph.Graph, v graph.VertexId) {
	// 		-> get all edges for this node
	edges := g.Edges[v]

	c.DownwardsGraph.AddVertex(g.Vertices[v])
	c.UpwardsGraph.AddVertex(g.Vertices[v])

	// 		-> for each edge:
	for _, edge := range edges {
		c.DownwardsGraph.AddVertex(g.Vertices[edge.Target])
		c.UpwardsGraph.AddVertex(g.Vertices[edge.Target])
		//-> if target is in the Contraction Order put the edge into the downwards graph.
		if slices.Contains(c.ContractionOrder, edge.Target) {
			c.UpwardsGraph.AddEdge(edge.Target, v, edge.Weight, edge.IsShortcut, edge.Via)
			c.DownwardsGraph.AddEdge(v, edge.Target, edge.Weight, edge.IsShortcut, edge.Via)
		} else {
			// 			-> else put it in the upwards graph
			c.DownwardsGraph.AddEdge(edge.Target, v, edge.Weight, edge.IsShortcut, edge.Via)
			c.UpwardsGraph.AddEdge(v, edge.Target, edge.Weight, edge.IsShortcut, edge.Via)
		}
		g.RemoveEdge(v, edge.Target)
		g.RemoveEdge(edge.Target, v)
	}
}

func Shortcuts(g *graph.Graph, v graph.VertexId, insertFlag bool) int {
	shortcutsFound := 0
	neighbors, _ := g.Neighbors(v)
	incidentEdges := g.Edges[v]

	for i := 0; i < len(neighbors)-1; i++ {
		u := neighbors[i]
		for j := i + 1; j < len(neighbors); j++ {
			w := neighbors[j]

			costViaV := float64(incidentEdges[u.Id].Weight) + float64(incidentEdges[w.Id].Weight)

			// Check if the path u->v->w is a shortest path. Bound the search by `costViaV`.
			_, shortestPathCost, _, _ := pathfinding.DijkstraShortestPath(g, u.Id, w.Id, costViaV /* bound */)
			if shortestPathCost < costViaV {
				continue // Path u->v->w is not a shortest path, so we ignore it.
			}

			// Check for an alternative path of the same length, ignoring v.
			// This search is also bounded by `costViaV`.
			_, _, _, err := pathfinding.DijkstraShortestPath(g, u.Id, w.Id, costViaV /* bound */, v /* ignored */)
			// A shortcut is needed only if the witness search fails to find a path within the bound.
			if err != nil {
				shortcutsFound++
				if insertFlag {
					ShortcutsAdded++
					cost := int(costViaV)
					addErr := g.AddEdge(u.Id, w.Id, cost, true, v)
					if addErr == graph.ErrEdgeAlreadyExists {
						g.UpdateEdge(u.Id, w.Id, cost, true, v)
						g.UpdateEdge(w.Id, u.Id, cost, true, v)
					} else {
						g.AddEdge(w.Id, u.Id, cost, true, v)
					}
				}
			}
		}
	}
	return shortcutsFound
}

// Ed is the number of shortcuts that would need to be added if contracting v minus the number of edges that would get contracted (degree of v).
// The priority combines the classic edge difference (shortcuts - degree) with a term that normalizes the number of
// shortcuts by the vertex's degree.
func Priority(g *graph.Graph, v graph.VertexId) float64 {
	degree, _ := g.Degree(v)
	shortcuts := Shortcuts(g, v, false)
	ed := EdgeDifference(g, v)

	priority := float64(ed) + (float64(shortcuts) / (float64(degree) + 1.0))

	return priority
}

func EdgeDifference(g *graph.Graph, v graph.VertexId) int {
	degree, _ := g.Degree(v)
	shortcuts := Shortcuts(g, v, false)
	return shortcuts - degree
}

// Query finds the shortest path between source and target using the CH.
// It performs a bidirectional Dijkstra search on the CH and then unpacks
// the resulting path to resolve any shortcuts.
func (c *ContractionHierarchies) Query(source, target graph.VertexId) ([]graph.VertexId, float64, int, error) {
	path, weight, nodesPopped, err := pathfinding.BiDirectionalDijkstraShortestPath(c.UpwardsGraph, c.DownwardsGraph, source, target)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("bidirectional Dijkstra failed: %w", err)
	}

	if len(path) == 0 {
		return []graph.VertexId{}, weight, 0, nil
	}

	unpackedPath, err := c.unpackPath(path)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to unpack path: %w", err)
	}

	return unpackedPath, weight, nodesPopped, nil
}

func (c *ContractionHierarchies) unpackPath(path []graph.VertexId) ([]graph.VertexId, error) {
	if len(path) < 2 {
		return path, nil
	}

	fullPath := []graph.VertexId{path[0]}

	for i := 0; i < len(path)-1; i++ {
		u, v := path[i], path[i+1]
		segment, err := c.unpackEdge(u, v)
		if err != nil {
			return nil, err
		}
		fullPath = append(fullPath, segment[1:]...)
	}

	return fullPath, nil
}

// unpackEdge recursively unpacks a single edge (u, v).
// If the edge is a shortcut, it finds the intermediate node and recursively
// unpacks the two new segments.
func (c *ContractionHierarchies) unpackEdge(u, v graph.VertexId) ([]graph.VertexId, error) {
	var edge graph.Edge
	var ok bool

	// The edge must exist in either the upwards or downwards graph.
	edge, ok = c.UpwardsGraph.Edges[u][v]
	if !ok {
		edge, ok = c.DownwardsGraph.Edges[u][v]
		if !ok {
			return nil, fmt.Errorf("no edge found between %d and %d in CH graphs", u, v)
		}
	}

	if !edge.IsShortcut {
		return []graph.VertexId{u, v}, nil
	}

	via := edge.Via
	path1, err := c.unpackEdge(u, via)
	if err != nil {
		return nil, err
	}
	path2, err := c.unpackEdge(via, v)
	if err != nil {
		return nil, err
	}

	// Combine the two unpacked sub-paths.
	// path1 is [u, ..., via], path2 is [via, ..., v].
	// append path2[1:] to path1 to get [u, ..., via, ..., v].
	return append(path1, path2[1:]...), nil
}

