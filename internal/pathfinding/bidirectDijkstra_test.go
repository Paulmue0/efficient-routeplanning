package pathfinding

import (
	"reflect" // Used for cleaner slice comparison
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

func TestBiDirectionalDijkstraShortestPath(t *testing.T) {
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
			wantCost: 5,
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
			target:   99, // Vertex not in graph
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
			var upGraph, downGraph *graph.Graph

			if tt.name == "no edges graph" {
				gNoEdges := graph.NewGraph()
				gNoEdges.AddVertex(graph.Vertex{Id: 0})
				gNoEdges.AddVertex(graph.Vertex{Id: 3})
				upGraph = gNoEdges
				downGraph = gNoEdges
			} else {
				upGraph = g
				downGraph = g
			}

			gotPath, gotCost, err := BiDirectionalDijkstraShortestPath(upGraph, downGraph, tt.source, tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("got path = %v, want %v", gotPath, tt.wantPath)
				t.Fatalf("BiDirectionalDijkstraShortestPath() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(gotPath, tt.wantPath) {
					t.Errorf("got path = %v, want %v", gotPath, tt.wantPath)
				}

				if gotCost != tt.wantCost {
					t.Errorf("got cost = %f, want %f", gotCost, tt.wantCost)
				}
			}
		})
	}
}
