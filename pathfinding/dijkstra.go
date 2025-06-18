package pathfinding

import (
	"errors"
	"math"

	collection "github.com/PaulMue0/efficient-routeplanning/Collection"
	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

var ErrTargetNotReachable = errors.New("target vertex not reachable from source")

func DijkstraShortestPath(g graph.Graph, source, target graph.VertexId) ([]graph.VertexId, error) {
	weights := initializeWeights(g, source)
	bestPredecessors := make(map[graph.VertexId]graph.VertexId)
	queue := initializePriorityQueue(g, weights)

	for queue.Len() > 0 {
		item := queue.Pop().(*collection.Item[graph.VertexId])
		vertex := item.Value

		if math.IsInf(weights[vertex], 1) {
			// Remaining vertices are unreachable from source
			break
		}
		if vertex == target {
			return buildPath(bestPredecessors, source, target)
		}

		for adjacent, edge := range g.Edges[vertex] {
			newWeight := weights[vertex] + float64(edge.Weight)
			if newWeight < weights[adjacent] {
				weights[adjacent] = newWeight
				bestPredecessors[adjacent] = vertex
				queue.UpdatePriority(adjacent, newWeight)
			}
		}
	}

	return nil, ErrTargetNotReachable
}

func initializeWeights(g graph.Graph, source graph.VertexId) map[graph.VertexId]float64 {
	weights := make(map[graph.VertexId]float64)
	for vertex := range g.Edges {
		if vertex == source {
			weights[vertex] = 0
		} else {
			weights[vertex] = math.Inf(1)
		}
	}
	return weights
}

func initializePriorityQueue(g graph.Graph, weights map[graph.VertexId]float64) *collection.PriorityQueue[graph.VertexId] {
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
