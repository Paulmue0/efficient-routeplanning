package pathfinding

import (
	"container/heap"
	"errors"
	"math"

	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
	collection "github.com/PaulMue0/efficient-routeplanning/pkg/collection/heap_gen"
)

var ErrTargetNotReachable = errors.New("target vertex not reachable from source")

func DijkstraShortestPath(g *graph.Graph, source, target graph.VertexId, bound float64, ignoredNode ...graph.VertexId) ([]graph.VertexId, float64, int, error) {
	// Only track distances for visited nodes - huge optimization
	distances := make(map[graph.VertexId]float64)
	distances[source] = 0
	bestPredecessors := make(map[graph.VertexId]graph.VertexId)

	// Only add source to queue initially - major optimization
	queue := collection.NewPriorityQueue[graph.VertexId]()
	queue.PushWithPriority(source, 0)

	// Track visited nodes to avoid re-processing
	visited := make(map[graph.VertexId]bool)
	nodesPopped := 0

	var nodeToIgnore graph.VertexId
	ignoreIsSet := len(ignoredNode) > 0
	if ignoreIsSet {
		nodeToIgnore = ignoredNode[0]
	}

	for queue.Len() > 0 {
		item := heap.Pop(queue).(*collection.Item[graph.VertexId])
		nodesPopped++
		vertex := queue.GetValue(item)
		cost := queue.GetPriority(item)

		// If already visited with better cost, skip
		if visited[vertex] {
			continue
		}
		visited[vertex] = true

		// If the shortest distance to the current node already exceeds the bound,
		// we know we can't find a path to the target within the limit.
		if cost >= bound {
			break
		}

		if ignoreIsSet && vertex == nodeToIgnore {
			continue
		}

		if vertex == target {
			path, err := buildPath(bestPredecessors, source, target)
			if err != nil {
				return nil, 0, nodesPopped, err
			}
			return path, distances[target], nodesPopped, nil
		}

		for adjacent, edge := range g.Edges[vertex] {
			if ignoreIsSet && adjacent == nodeToIgnore {
				continue
			}

			if visited[adjacent] {
				continue
			}

			newWeight := cost + float64(edge.Weight)

			// Early pruning - don't explore paths that exceed bound
			if newWeight >= bound {
				continue
			}

			if oldDist, exists := distances[adjacent]; !exists || newWeight < oldDist {
				distances[adjacent] = newWeight
				bestPredecessors[adjacent] = vertex
				queue.PushWithPriority(adjacent, newWeight)
			}
		}
	}

	return nil, 0, nodesPopped, ErrTargetNotReachable
}

// WitnessSearch is an optimized version for contraction hierarchies
// Returns true if a witness path exists (path not using ignored node within bound)
func WitnessSearch(g *graph.Graph, source, target graph.VertexId, bound float64, ignoredNode graph.VertexId) bool {
	distances := make(map[graph.VertexId]float64)
	distances[source] = 0

	queue := collection.NewPriorityQueue[graph.VertexId]()
	queue.PushWithPriority(source, 0)

	visited := make(map[graph.VertexId]bool)

	for queue.Len() > 0 {
		item := heap.Pop(queue).(*collection.Item[graph.VertexId])
		vertex := queue.GetValue(item)
		cost := queue.GetPriority(item)

		if visited[vertex] {
			continue
		}
		visited[vertex] = true

		// Stop if we exceed bound
		if cost >= bound {
			return false
		}

		// Found witness path
		if vertex == target {
			return true
		}

		for adjacent, edge := range g.Edges[vertex] {
			// Skip ignored node
			if adjacent == ignoredNode {
				continue
			}

			if visited[adjacent] {
				continue
			}

			newWeight := cost + float64(edge.Weight)

			// Prune paths that exceed bound
			if newWeight >= bound {
				continue
			}

			if oldDist, exists := distances[adjacent]; !exists || newWeight < oldDist {
				distances[adjacent] = newWeight
				queue.PushWithPriority(adjacent, newWeight)
			}
		}
	}

	return false // No witness path found
}

func initializeWeights(g *graph.Graph, source graph.VertexId) map[graph.VertexId]float64 {
	weights := make(map[graph.VertexId]float64)
	for vertex := range g.Vertices {
		if vertex == source {
			weights[vertex] = 0
		} else {
			weights[vertex] = math.Inf(1)
		}
	}
	return weights
}

func initializePriorityQueue(weights map[graph.VertexId]float64) *collection.PriorityQueue[graph.VertexId] {
	queue := collection.NewPriorityQueue[graph.VertexId]()
	for vertex, weight := range weights {
		queue.PushWithPriority(vertex, weight)
	}
	return queue
}

func buildPath(predecessors map[graph.VertexId]graph.VertexId, source, target graph.VertexId) ([]graph.VertexId, error) {
	path := []graph.VertexId{target}
	current := target

	for current != source {
		prev, exists := predecessors[current]
		if !exists {
			return nil, ErrTargetNotReachable
		}
		current = prev
		path = append([]graph.VertexId{current}, path...)
	}

	return path, nil
}
