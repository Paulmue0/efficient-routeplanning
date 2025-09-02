package pathfinding

import (
	"container/heap"
	"math"

	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
	collection "github.com/PaulMue0/efficient-routeplanning/pkg/collection/heap_gen"
)

// searchContext holds the state for a single direction of a Dijkstra search.
// It includes the graph being traversed, a priority queue for nodes to visit,
// maps for distances from the source, and predecessors to reconstruct the path.
type searchContext struct {
	g     *graph.Graph
	pq    *collection.PriorityQueue[graph.VertexId]
	dists map[graph.VertexId]float64
	preds map[graph.VertexId]graph.VertexId
}

// newSearchContext initializes and returns a new search context for a given graph and start node.
// It sets up the distance map with infinity for all nodes except the start node (which is set to 0)
// and prepares the priority queue.
func newSearchContext(g *graph.Graph, startNode graph.VertexId) *searchContext {
	dists := initializeWeights(g, startNode)
	pq := initializePriorityQueue(dists)
	return &searchContext{
		g:     g,
		pq:    pq,
		dists: dists,
		preds: make(map[graph.VertexId]graph.VertexId),
	}
}

// processNextNode extracts the highest-priority node from the queue, relaxes its edges,
// and checks for a potential meeting point with the opposite search direction.
// It updates the shortest path length and the meeting node if a shorter path is found.
func (sc *searchContext) processNextNode(
	oppositeDists map[graph.VertexId]float64,
	shortestPathLength *float64,
	meetNode *graph.VertexId,
) {
	item := heap.Pop(sc.pq).(*collection.Item[graph.VertexId])
	vertex := sc.pq.GetValue(item)
	cost := sc.pq.GetPriority(item)

	// Check if the current node has been reached by the other search.
	// If so, a potential path has been found.
	if oppositeDist, found := oppositeDists[vertex]; found && !math.IsInf(oppositeDist, 1) {
		if potentialPathLength := cost + oppositeDist; potentialPathLength < *shortestPathLength {
			*shortestPathLength = potentialPathLength
			*meetNode = vertex
		}
	}

	// Relax outgoing edges.
	for adjacent, edge := range sc.g.Edges[vertex] {
		newWeight := cost + float64(edge.Weight)
		if newWeight < sc.dists[adjacent] {
			sc.dists[adjacent] = newWeight
			sc.preds[adjacent] = vertex
			sc.pq.UpdatePriority(adjacent, newWeight)
		}
	}
}

// BiDirectionalDijkstraShortestPath finds the shortest path between a source and target node
// in a graph using a bidirectional Dijkstra's algorithm. It runs two searches simultaneously:
// one forward from the source on upGraph and one backward from the target on downGraph.
// For Contraction Hierarchies, upGraph contains only upward edges, and downGraph contains only
// downward edges. The search terminates when the sum of minimum distances from both search
// frontiers exceeds the length of the best path found so far.
// It returns the path as a slice of vertex IDs, the total path weight, and an error if no
// path is found.
func BiDirectionalDijkstraShortestPath(upGraph *graph.Graph, downGraph *graph.Graph, source, target graph.VertexId) ([]graph.VertexId, float64, error) {
	if source == target {
		if _, ok := upGraph.Edges[source]; ok {
			return []graph.VertexId{source}, 0, nil
		}
		return nil, 0, ErrTargetNotReachable
	}

	fwdSearch := newSearchContext(upGraph, source)
	bwdSearch := newSearchContext(upGraph, target)

	currentShortestPath := math.Inf(1)
	var meetNode graph.VertexId

	for fwdSearch.pq.Len() > 0 && bwdSearch.pq.Len() > 0 {
		fwdMinDist := fwdSearch.pq.GetPriority(fwdSearch.pq.Peek())
		bwdMinDist := bwdSearch.pq.GetPriority(bwdSearch.pq.Peek())

		// Termination condition: if the sum of the smallest distances in both queues
		// is greater than or equal to the current shortest path, no shorter path can be found.
		// This check is only performed if a path has already been found.
		if !math.IsInf(currentShortestPath, 1) && fwdMinDist+bwdMinDist >= currentShortestPath {
			break
		}

		// Process the node from the search direction with the smaller minimum distance.
		if fwdMinDist <= bwdMinDist {
			fwdSearch.processNextNode(bwdSearch.dists, &currentShortestPath, &meetNode)
		} else {
			bwdSearch.processNextNode(fwdSearch.dists, &currentShortestPath, &meetNode)
		}
	}

	// One of the searches may be exhausted. Continue with the other until its priority queue
	// is empty or the minimum distance is greater than the current shortest path.
	for fwdSearch.pq.Len() > 0 && (math.IsInf(currentShortestPath, 1) || fwdSearch.pq.GetPriority(fwdSearch.pq.Peek()) < currentShortestPath) {
		fwdSearch.processNextNode(bwdSearch.dists, &currentShortestPath, &meetNode)
	}
	for bwdSearch.pq.Len() > 0 && (math.IsInf(currentShortestPath, 1) || bwdSearch.pq.GetPriority(bwdSearch.pq.Peek()) < currentShortestPath) {
		bwdSearch.processNextNode(fwdSearch.dists, &currentShortestPath, &meetNode)
	}

	if math.IsInf(currentShortestPath, 1) {
		return nil, 0, ErrTargetNotReachable
	}

	pathFwd, errFwd := buildPath(fwdSearch.preds, source, meetNode)
	if errFwd != nil {
		return nil, 0, errFwd
	}

	pathBwdReversed, errBwd := buildPath(bwdSearch.preds, target, meetNode)
	if errBwd != nil {
		return nil, 0, errBwd
	}

	// Reverse the backward path to get the correct order from meetNode to target.
	pathBwd := make([]graph.VertexId, len(pathBwdReversed))
	for i, j := 0, len(pathBwdReversed)-1; j >= 0; i, j = i+1, j-1 {
		pathBwd[i] = pathBwdReversed[j]
	}

	// Combine the forward and backward paths, excluding the duplicated meetNode.
	return append(pathFwd, pathBwd[1:]...), currentShortestPath, nil
}
