package pathfinding

import (
	"fmt"
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

var (
	mockGraph       = graph.NewGraph()
	mockRoadNetwork = graph.RoadNetwork{NumNodes: 4, NumEdges: 5, Network: mockGraph}
)

func TestDijkstra(t *testing.T) {
	// Define vertices (nodes)
	v0 := graph.Vertex{Id: 0, Lat: 48.667421, Lon: 9.244557}
	v1 := graph.Vertex{Id: 1, Lat: 48.667273, Lon: 9.244867}
	v2 := graph.Vertex{Id: 2, Lat: 48.667598, Lon: 9.244326}
	v3 := graph.Vertex{Id: 3, Lat: 48.667019, Lon: 9.245514}

	// Add vertices to the graph
	mockGraph.AddVertex(v0)
	mockGraph.AddVertex(v1)
	mockGraph.AddVertex(v2)
	mockGraph.AddVertex(v3)

	// Add edges (with weights)
	mockGraph.AddEdge(v0.Id, v1.Id, 2) // v0 → v1 (cost 2)
	mockGraph.AddEdge(v1.Id, v2.Id, 4) // v1 → v2 (cost 4)
	mockGraph.AddEdge(v0.Id, v2.Id, 1) // v0 → v2 (cost 1)
	mockGraph.AddEdge(v2.Id, v3.Id, 7) // v2 → v3 (cost 7)
	mockGraph.AddEdge(v1.Id, v3.Id, 3) // v1 → v3 (cost 3)

	// Optionally add reverse edges for undirected graph
	mockGraph.AddEdge(v1.Id, v0.Id, 2)
	mockGraph.AddEdge(v2.Id, v1.Id, 4)
	mockGraph.AddEdge(v2.Id, v0.Id, 1)
	mockGraph.AddEdge(v3.Id, v2.Id, 7)
	mockGraph.AddEdge(v3.Id, v1.Id, 3)

	got, _ := ShortestPath(*mockGraph, v0.Id, v3.Id)
	fmt.Println(got)
	// Expected shortest path: v0 → v1 → v3 with cost = 5
}
