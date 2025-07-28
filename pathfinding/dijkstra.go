package pathfinding

import (
	"container/heap"
	"errors"
	"math"

	collection "github.com/PaulMue0/efficient-routeplanning/Collection"
	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

var ErrTargetNotReachable = errors.New("target vertex not reachable from source")

func DijkstraShortestPath(g *graph.Graph, source, target graph.VertexId) ([]graph.VertexId, float64, error) {
	weights := initializeWeights(g, source)
	bestPredecessors := make(map[graph.VertexId]graph.VertexId)
	queue := initializePriorityQueue(weights)

	for queue.Len() > 0 {
		item := heap.Pop(queue).(*collection.Item[graph.VertexId])
		vertex := queue.GetValue(item)

		if math.IsInf(weights[vertex], 1) {
			break
		}
		if vertex == target {
			path, err := buildPath(bestPredecessors, source, target)
			if err != nil {
				return nil, 0, err
			}
			return path, weights[target], nil
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
	return nil, 0, ErrTargetNotReachable
}

func initializeWeights(g *graph.Graph, source graph.VertexId) map[graph.VertexId]float64 {
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
