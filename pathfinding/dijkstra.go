package pathfinding

import (
	"errors"
	"math"

	collection "github.com/PaulMue0/efficient-routeplanning/Collection"
	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

var ErrTargetNotReachable = errors.New("target vertex not reachable from source")

func ShortestPath(g graph.Graph, source, target graph.VertexId) ([]graph.VertexId, error) {
	weights := make(map[graph.VertexId]float64)
	visited := make(map[graph.VertexId]bool)

	weights[source] = 0
	visited[target] = true

	queue := collection.NewPriorityQueue[graph.VertexId]()
	adjacencyMap := g.Edges

	for vertex := range adjacencyMap {
		if vertex != source {
			weights[vertex] = math.Inf(1)
			visited[vertex] = false
		}
		queue.PushWithPriority(vertex, weights[vertex])
	}

	bestPredecessors := make(map[graph.VertexId]graph.VertexId)

	for queue.Len() > 0 {
		item := queue.Pop().(*collection.Item[graph.VertexId])
		vertex := item.Value

		hasInfiniteWeight := math.IsInf(weights[vertex], 1)

		if hasInfiniteWeight {
			// Remaining vertices are unreachable
			break
		}
		if vertex == target {
			break // or return the path immediately
		}

		for adjacency, edge := range adjacencyMap[vertex] {
			edgeWeight := edge.Weight
			newWeight := weights[vertex] + float64(edgeWeight)

			if newWeight < weights[adjacency] {
				weights[adjacency] = newWeight
				bestPredecessors[adjacency] = vertex
				queue.UpdatePriority(adjacency, newWeight)
			}
		}
	}

	path := []graph.VertexId{target}
	current := target

	for current != source {
		prev, ok := bestPredecessors[current]
		if !ok {
			return nil, ErrTargetNotReachable
		}
		current = prev
		path = append([]graph.VertexId{current}, path...)
	}

	return path, nil
}

// 1   function Dijkstra(Graph, source):
// 2       Q ← Queue storing vertex priority
// 3
// 4       dist[source] ← 0                          // Initialization
// 5       Q.add_with_priority(source, 0)            // associated priority equals dist[·]
// 6
// 7       for each vertex v in Graph.Vertices:
// 8           if v ≠ source
// 9               prev[v] ← UNDEFINED               // Predecessor of v
// 10              dist[v] ← INFINITY                // Unknown distance from source to v
// 11              Q.add_with_priority(v, INFINITY)
// 12
// 13
// 14      while Q is not empty:                     // The main loop
// 15          u ← Q.extract_min()                   // Remove and return best vertex
// 16          for each arc (u, v) :                 // Go through all v neighbors of u
// 17              alt ← dist[u] + Graph.Edges(u, v)
// 18              if alt < dist[v]:
// 19                  prev[v] ← u
// 20                  dist[v] ← alt
// 21                  Q.decrease_priority(v, alt)
// 22
// 23      return (dist, prev)
