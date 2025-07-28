package pathfinding

import (
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

func createTestGraph() *graph.Graph {
	g := graph.NewGraph()

	v0 := graph.Vertex{Id: 0, Lat: 48.667421, Lon: 9.244557}
	v1 := graph.Vertex{Id: 1, Lat: 48.667273, Lon: 9.244867}
	v2 := graph.Vertex{Id: 2, Lat: 48.667598, Lon: 9.244326}
	v3 := graph.Vertex{Id: 3, Lat: 48.667019, Lon: 9.245514}

	// Add vertices
	g.AddVertex(v0)
	g.AddVertex(v1)
	g.AddVertex(v2)
	g.AddVertex(v3)

	// Add edges (bidirectional)
	g.AddEdge(v0.Id, v1.Id, 2, false, -1)
	g.AddEdge(v1.Id, v0.Id, 2, false, -1)

	g.AddEdge(v1.Id, v2.Id, 4, false, -1)
	g.AddEdge(v2.Id, v1.Id, 4, false, -1)

	g.AddEdge(v0.Id, v2.Id, 1, false, -1)
	g.AddEdge(v2.Id, v0.Id, 1, false, -1)

	g.AddEdge(v2.Id, v3.Id, 7, false, -1)
	g.AddEdge(v3.Id, v2.Id, 7, false, -1)

	g.AddEdge(v1.Id, v3.Id, 3, false, -1)
	g.AddEdge(v3.Id, v1.Id, 3, false, -1)

	return g
}

func pathCost(g *graph.Graph, path []graph.VertexId) float64 {
	var cost float64
	for i := range path[:len(path)-1] {
		from := path[i]
		to := path[i+1]

		found := false
		for adj, edge := range g.Edges[from] {
			if adj == to {
				cost += float64(edge.Weight)
				found = true
				break
			}
		}
		if !found {
			// Invalid path edge
			return -1
		}
	}
	return cost
}

func TestDijkstraShortestPath(t *testing.T) {
	g := createTestGraph()

	tests := []struct {
		name     string
		source   graph.VertexId
		target   graph.VertexId
		wantPath []graph.VertexId
		wantErr  bool
		wantCost float64
	}{
		{
			name:     "shortest path v0 to v3",
			source:   0,
			target:   3,
			wantPath: []graph.VertexId{0, 1, 3},
			wantErr:  false,
			wantCost: 5, // 2 + 3
		},
		{
			name:     "shortest path v0 to v2",
			source:   0,
			target:   2,
			wantPath: []graph.VertexId{0, 2},
			wantErr:  false,
			wantCost: 1,
		},
		{
			name:     "source equals target",
			source:   1,
			target:   1,
			wantPath: []graph.VertexId{1},
			wantErr:  false,
			wantCost: 0,
		},
		{
			name:     "target unreachable",
			source:   0,
			target:   99, // vertex not in graph
			wantPath: nil,
			wantErr:  true,
		},
		{
			name:     "no edges graph",
			source:   0,
			target:   3,
			wantPath: nil,
			wantErr:  true,
			wantCost: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For "no edges graph" test case, create a graph with no edges
			var graphToUse *graph.Graph
			if tt.name == "no edges graph" {
				graphToUse = graph.NewGraph()
				graphToUse.AddVertex(graph.Vertex{Id: 0})
				graphToUse.AddVertex(graph.Vertex{Id: 3})
			} else {
				graphToUse = g
			}

			gotPath, _, err := DijkstraShortestPath(graphToUse, tt.source, tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DijkstraShortestPath() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Check path correctness
				if len(gotPath) != len(tt.wantPath) {
					t.Errorf("got path length = %d, want %d", len(gotPath), len(tt.wantPath))
				} else {
					for i := range gotPath {
						if gotPath[i] != tt.wantPath[i] {
							t.Errorf("got path[%d] = %d, want %d", i, gotPath[i], tt.wantPath[i])
						}
					}
				}

				// Check cost correctness
				gotCost := pathCost(graphToUse, gotPath)
				if gotCost != tt.wantCost {
					t.Errorf("got cost = %f, want %f", gotCost, tt.wantCost)
				}
			}
		})
	}
}
