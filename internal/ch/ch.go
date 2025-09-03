package ch

import (
	"container/heap"
	"fmt"
	"slices"
	"sync"

	pathfinding "github.com/PaulMue0/efficient-routeplanning/internal/pathfinding"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
	collection "github.com/PaulMue0/efficient-routeplanning/pkg/collection/heap_gen"
)

// ShortcutsAdded is a global counter for the number of shortcuts added during preprocessing.
// It is used for experimental analysis.
var ShortcutsAdded = 0

// ContractionHierarchies represents the data structure for contraction hierarchies.
// It contains the original graph, the upward and downward graphs built during preprocessing,
// the contraction order of vertices, and a priority queue for selecting vertices to contract.
type ContractionHierarchies struct {
	NumShortcutsAdded int
	ContractionOrder  []graph.VertexId
	Priorities        *collection.PriorityQueue[graph.VertexId]
	UpwardsGraph      *graph.Graph
	DownwardsGraph    *graph.Graph
}

// NewContractionHierarchies creates and initializes a new ContractionHierarchies struct.
func NewContractionHierarchies() *ContractionHierarchies {
	co := make([]graph.VertexId, 0)
	ed := collection.NewPriorityQueue[graph.VertexId]()
	ug := graph.NewGraph()
	dg := graph.NewGraph()

	return &ContractionHierarchies{0, co, ed, ug, dg}
}

// Preprocess prepares the graph for fast queries by contracting vertices in an optimized order.
// It iteratively finds batches of independent vertices and contracts them in parallel.
func (c *ContractionHierarchies) Preprocess(g *graph.Graph) {
	const batchSize = 128
	c.InitializePriority(g)

	for len(g.Vertices) > 0 {
		independentSet := c.findIndependentSet(g, batchSize)

		if len(independentSet) == 0 {
			if c.Priorities.Len() > 0 {
				item := heap.Pop(c.Priorities).(*collection.Item[graph.VertexId])
				v := c.Priorities.GetValue(item)
				if _, ok := g.Vertices[v]; ok {
					independentSet = []graph.VertexId{v}
				} else {
					continue
				}
			} else {
				break
			}
		}

		allNeighbors := make(map[graph.VertexId]struct{})
		neighborsMap := make(map[graph.VertexId][]graph.Vertex)
		for _, v := range independentSet {
			neighbors, _ := g.Neighbors(v)
			neighborsMap[v] = neighbors
			for _, neighbor := range neighbors {
				allNeighbors[neighbor.Id] = struct{}{}
			}
		}

		c.contractBatch(g, independentSet, neighborsMap)

		c.recomputeBatchNeighborPriorities(g, allNeighbors)
	}
}

// findIndependentSet selects a set of vertices that can be contracted in parallel without causing conflicts.
// Vertices are considered independent if they are not adjacent and do not share any common neighbors.
// It prioritizes vertices with a lower contraction priority (e.g., smaller edge difference).
func (c *ContractionHierarchies) findIndependentSet(g *graph.Graph, batchSize int) []graph.VertexId {
	independentSet := make([]graph.VertexId, 0, batchSize)
	nodesInSet := make(map[graph.VertexId]struct{})
	neighborsOfSet := make(map[graph.VertexId]struct{})

	tempPopped := make([]*collection.Item[graph.VertexId], 0)

	for len(independentSet) < batchSize && c.Priorities.Len() > 0 {
		item := heap.Pop(c.Priorities).(*collection.Item[graph.VertexId])
		v := c.Priorities.GetValue(item)

		if _, ok := g.Vertices[v]; !ok {
			continue // already contracted
		}

		isIndependent := true
		if _, exists := neighborsOfSet[v]; exists {
			isIndependent = false
		}

		if isIndependent {
			neighbors, _ := g.Neighbors(v)
			for _, neighbor := range neighbors {
				if _, exists := nodesInSet[neighbor.Id]; exists {
					isIndependent = false
					break
				}
				if _, exists := neighborsOfSet[neighbor.Id]; exists {
					isIndependent = false
					break
				}
			}
		}

		if isIndependent {
			independentSet = append(independentSet, v)
			nodesInSet[v] = struct{}{}
			neighbors, _ := g.Neighbors(v)
			for _, neighbor := range neighbors {
				neighborsOfSet[neighbor.Id] = struct{}{}
			}
		} else {
			tempPopped = append(tempPopped, item)
		}
	}

	for _, item := range tempPopped {
		heap.Push(c.Priorities, item)
	}

	return independentSet
}

// contractBatch contracts a given set of independent vertices.
// The process is parallelized by first calculating all necessary shortcuts concurrently.
// Then, all graph modifications (adding shortcuts, removing vertices) are applied sequentially
// to maintain data consistency.
// Protects global counter ShortcutsAdded
var shortcutsMu sync.Mutex

func (c *ContractionHierarchies) contractBatch(
	g *graph.Graph,
	vertices []graph.VertexId,
	neighborsMap map[graph.VertexId][]graph.Vertex,
) {
	type shortcut struct {
		from, to, via graph.VertexId
		weight        int
	}

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		shortcuts []shortcut
	)

	// --- Phase 1: find shortcuts in parallel ---
	for _, v := range vertices {
		wg.Add(1)
		go func(vertexId graph.VertexId) {
			defer wg.Done()
			neighbors := neighborsMap[vertexId]
			incidentEdges := g.Edges[vertexId]

			for i := 0; i < len(neighbors)-1; i++ {
				u := neighbors[i]
				for j := i + 1; j < len(neighbors); j++ {
					w := neighbors[j]

					costViaV := float64(incidentEdges[u.Id].Weight) + float64(incidentEdges[w.Id].Weight)

					// Check if existing path beats the shortcut
					_, shortestPathCost, _, _ := pathfinding.DijkstraShortestPath(g, u.Id, w.Id, costViaV)
					if shortestPathCost < costViaV {
						continue
					}

					// If ignoring vertexId breaks connectivity, we need a shortcut
					_, _, _, err := pathfinding.DijkstraShortestPath(g, u.Id, w.Id, costViaV, vertexId)
					if err != nil {
						mu.Lock()
						shortcuts = append(shortcuts, shortcut{
							from: u.Id, to: w.Id, via: vertexId, weight: int(costViaV),
						})
						mu.Unlock()
					}
				}
			}
		}(v)
	}
	wg.Wait()

	// --- Phase 2: apply graph modifications sequentially ---
	for _, sc := range shortcuts {
		c.NumShortcutsAdded++

		shortcutsMu.Lock()
		ShortcutsAdded++
		shortcutsMu.Unlock()

		cost := sc.weight
		addErr := g.AddEdge(sc.from, sc.to, cost, true, sc.via)
		if addErr == graph.ErrEdgeAlreadyExists {
			existingEdge := g.Edges[sc.from][sc.to]
			if cost < existingEdge.Weight {
				g.UpdateEdge(sc.from, sc.to, cost, true, sc.via)
				g.UpdateEdge(sc.to, sc.from, cost, true, sc.via)
			}
		} else {
			g.AddEdge(sc.to, sc.from, cost, true, sc.via)
		}
	}

	// --- Phase 3: contract vertices ---
	for _, v := range vertices {
		c.ContractionOrder = append(c.ContractionOrder, v)
		c.InsertInUpwardsOrDownwardsGraph(g, v)
		if err := g.RemoveVertex(v); err != nil {
			panic(fmt.Sprintf("critical error removing vertex %v: %v", v, err))
		}
	}
}

// InitializePriority computes the initial priority for every vertex in the graph and
// populates the priority queue.
func (c *ContractionHierarchies) InitializePriority(g *graph.Graph) {
	for _, v := range g.Vertices {
		c.Priorities.PushWithPriority(v.Id, c.Priority(g, v.Id))
	}
}

// recomputeBatchNeighborPriorities recalculates priorities for all neighbors
// of the contracted batch in parallel and applies updates sequentially.
func (c *ContractionHierarchies) recomputeBatchNeighborPriorities(
	g *graph.Graph, allNeighbors map[graph.VertexId]struct{},
) {
	var wg sync.WaitGroup
	results := make(chan struct {
		id       graph.VertexId
		priority float64
	}, len(allNeighbors))

	for neighborId := range allNeighbors {
		if _, ok := g.Vertices[neighborId]; !ok {
			continue // skip already contracted
		}
		wg.Add(1)
		go func(n graph.VertexId) {
			defer wg.Done()
			prio := c.Priority(g, n)
			results <- struct {
				id       graph.VertexId
				priority float64
			}{id: n, priority: prio}
		}(neighborId)
	}

	wg.Wait()
	close(results)

	for r := range results {
		c.Priorities.UpdatePriority(r.id, r.priority)
	}
}

// recomputeNeighborPriorities recalculates priorities for a set of neighbor vertices
// in parallel and applies updates to the priority queue sequentially.
func (c *ContractionHierarchies) recomputeNeighborPriorities(g *graph.Graph, neighbors map[graph.VertexId]struct{}) {
	var wg sync.WaitGroup
	results := make(chan struct {
		id       graph.VertexId
		priority float64
	}, len(neighbors))

	for neighborId := range neighbors {
		if _, ok := g.Vertices[neighborId]; !ok {
			continue // skip already removed vertices
		}
		wg.Add(1)
		go func(n graph.VertexId) {
			defer wg.Done()
			prio := c.Priority(g, n)
			results <- struct {
				id       graph.VertexId
				priority float64
			}{id: n, priority: prio}
		}(neighborId)
	}

	wg.Wait()
	close(results)

	for r := range results {
		c.Priorities.UpdatePriority(r.id, r.priority)
	}
}

// Contract contracts a single vertex v. This involves adding shortcuts between its neighbors
// to preserve shortest path distances, and then removing v from the graph.
// The contracted vertex is added to the contraction order.
func (c *ContractionHierarchies) Contract(g *graph.Graph, v graph.VertexId) {
	c.Shortcuts(g, v, true)
	c.ContractionOrder = append(c.ContractionOrder, v)
	c.InsertInUpwardsOrDownwardsGraph(g, v)

	if err := g.RemoveVertex(v); err != nil {
		panic(fmt.Sprintf("critical error removing vertex %v: %v.\n Edges: %v,\n Vertices %v", v, err, g.Edges[v], g.Vertices))
	}
}

// InsertInUpwardsOrDownwardsGraph moves the edges of a contracted vertex v into the
// upward and downward graphs of the contraction hierarchy. An edge is added to the
// upward graph if the target vertex has a higher contraction order, and to the
// downward graph otherwise.
func (c *ContractionHierarchies) InsertInUpwardsOrDownwardsGraph(g *graph.Graph, v graph.VertexId) {
	edges := g.Edges[v]

	c.DownwardsGraph.AddVertex(g.Vertices[v])
	c.UpwardsGraph.AddVertex(g.Vertices[v])

	for _, edge := range edges {
		c.DownwardsGraph.AddVertex(g.Vertices[edge.Target])
		c.UpwardsGraph.AddVertex(g.Vertices[edge.Target])
		if slices.Contains(c.ContractionOrder, edge.Target) {
			c.UpwardsGraph.AddEdge(edge.Target, v, edge.Weight, edge.IsShortcut, edge.Via)
			c.DownwardsGraph.AddEdge(v, edge.Target, edge.Weight, edge.IsShortcut, edge.Via)
		} else {
			c.DownwardsGraph.AddEdge(edge.Target, v, edge.Weight, edge.IsShortcut, edge.Via)
			c.UpwardsGraph.AddEdge(v, edge.Target, edge.Weight, edge.IsShortcut, edge.Via)
		}
		g.RemoveEdge(v, edge.Target)
		g.RemoveEdge(edge.Target, v)
	}
}

// Shortcuts calculates the number of shortcuts required if vertex v were to be contracted.
// A shortcut is an edge added between two neighbors of v to preserve shortest path distances.
// If the insertFlag is true, it adds the necessary shortcuts to the graph.
func (c *ContractionHierarchies) Shortcuts(g *graph.Graph, v graph.VertexId, insertFlag bool) int {
	shortcutsFound := 0
	neighbors, _ := g.Neighbors(v)
	incidentEdges := g.Edges[v]

	for i := 0; i < len(neighbors)-1; i++ {
		u := neighbors[i]
		for j := i + 1; j < len(neighbors); j++ {
			w := neighbors[j]

			costViaV := float64(incidentEdges[u.Id].Weight) + float64(incidentEdges[w.Id].Weight)

			_, shortestPathCost, _, _ := pathfinding.DijkstraShortestPath(g, u.Id, w.Id, costViaV)
			if shortestPathCost < costViaV {
				continue
			}

			_, _, _, err := pathfinding.DijkstraShortestPath(g, u.Id, w.Id, costViaV, v)
			if err != nil {
				shortcutsFound++
				if insertFlag {
					c.NumShortcutsAdded++
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

// Priority calculates the contraction priority for a vertex v. The priority is a heuristic
// used to decide the order of contraction. It is typically based on the edge difference,
// which is the number of shortcuts added minus the number of edges removed.
func (c *ContractionHierarchies) Priority(g *graph.Graph, v graph.VertexId) float64 {
	degree, _ := g.Degree(v)
	shortcuts := c.Shortcuts(g, v, false)
	ed := shortcuts - degree

	priority := float64(ed) + (float64(shortcuts) / (float64(degree) + 1.0))

	return priority
}

// Query finds the shortest path between a source and a target vertex using the preprocessed
// contraction hierarchy. It performs a bidirectional Dijkstra search on the upward and downward
// graphs and then unpacks the resulting path to resolve any shortcuts.
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

// unpackPath reconstructs the full shortest path from a path that may contain shortcuts.
// It iterates through the path segments and recursively unpacks any shortcut edges.
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

// unpackEdge recursively unpacks a single edge (u, v). If the edge is a shortcut,
// it finds the intermediate vertex and recursively unpacks the two resulting sub-paths.
func (c *ContractionHierarchies) unpackEdge(u, v graph.VertexId) ([]graph.VertexId, error) {
	var edge graph.Edge
	var ok bool

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

	return append(path1, path2[1:]...), nil
}
