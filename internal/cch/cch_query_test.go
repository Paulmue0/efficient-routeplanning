package cch

import (
	"errors"
	"log"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/PaulMue0/efficient-routeplanning/internal/parser"
	"github.com/PaulMue0/efficient-routeplanning/internal/pathfinding"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

func TestOsm1Query(t *testing.T) {
	t.Run("osm1 from 1 to 5", func(t *testing.T) {
		name := "osm1.txt"
		dataDir := "../../data/RoadNetworks"
		fileSystem := os.DirFS(dataDir)
		network, err := parser.NewNetworkFromFS(fileSystem, name)
		if err != nil {
			log.Fatalf("Failed to load graph: %v", err)
		}
		log.Printf("File: %s, NumNodes: %d, NumEdges: %d", name, network.NumNodes, network.NumEdges)

		// Preprocess CCH
		cchInst := NewCCH()
		log.Println("Starting CCH preprocessing...")
		start := time.Now()
		err = cchInst.Preprocess(network.Network, "../../data/kaHIP/osm1.ordering")
		if err != nil {
			log.Fatalf("CCH preprocessing failed: %v", err)
		}
		duration := time.Since(start)
		log.Printf("Finished CCH preprocessing in %s", duration)
		cchInst.Customize(network.Network)

		path, dist, _, err := cchInst.Query(1, 5)
		if slices.Equal(path, []graph.VertexId{1, 2, 3, 4, 5}) {
			t.Errorf("wrong path: Got %q Expected: %q", path, []graph.VertexId{1, 2, 3, 4, 5})
		}
		if dist != 5 {
			t.Errorf("wrong path length. Got %f Expected %f", dist, 5.0)
		}
		if err != nil {

			t.Error(err)
		}
	})
}
func TestCCHQuery(t *testing.T) {
	tests := []struct {
		name           string
		vertices       []graph.VertexId
		edges          [][3]int
		ordering       string
		source         graph.VertexId
		target         graph.VertexId
		expectedPath   []graph.VertexId
		expectedWeight float64
		expectError    bool
	}{
		{
			name:     "simple path with one shortcut",
			vertices: []graph.VertexId{0, 1, 2},
			edges: [][3]int{
				{0, 1, 1},
				{0, 2, 1},
				{1, 2, 5}, // This edge is more expensive than the path 1->0->2
			},
			// Contract 0 first to create a shortcut between 1 and 2.
			ordering:       "3\n1 1\n2 2\n3 3\n", // Order: 0, 1, 2
			source:         1,
			target:         2,
			expectedPath:   []graph.VertexId{1, 0, 2},
			expectedWeight: 2,
		},
		{
			name:     "path with no shortcuts",
			vertices: []graph.VertexId{0, 1, 2},
			edges: [][3]int{
				{0, 1, 1},
				{1, 2, 1},
			},
			ordering:       "3\n1 1\n2 2\n3 3\n", // Order: 0, 1, 2
			source:         0,
			target:         2,
			expectedPath:   []graph.VertexId{0, 1, 2},
			expectedWeight: 2,
		},
		{
			name: "more complex graph with multiple shortcuts",
			//      4 --(1)-- 5
			//     /           \
			// (1)/             \(1)
			//   0 ---(10)--- 1 ---(10)--- 2 ---(10)--- 3
			vertices: []graph.VertexId{0, 1, 2, 3, 4, 5},
			edges: [][3]int{
				{0, 1, 10}, {1, 2, 10}, {2, 3, 10}, // long path
				{0, 4, 1}, {4, 5, 1}, {5, 1, 1}, // shortcut path for 0->1
			},
			// Contract 4, 5 first to create a shortcut for 0->1
			ordering:       "6\n1 5\n2 6\n3 1\n4 2\n5 3\n6 4\n", // Order: 4, 5, 0, 1, 2, 3
			source:         0,
			target:         3,
			expectedPath:   []graph.VertexId{0, 4, 5, 1, 2, 3},
			expectedWeight: 23, // 1+1+1 + 10 + 10
		},
		{
			name:           "no path between source and target",
			vertices:       []graph.VertexId{0, 1, 2, 3},
			edges:          [][3]int{{0, 1, 1}, {2, 3, 1}},
			ordering:       "4\n1 1\n2 2\n3 3\n4 4\n",
			source:         0,
			target:         3,
			expectedPath:   nil,
			expectedWeight: 0,
			expectError:    true,
		},
		{
			name:     "source equals target",
			vertices: []graph.VertexId{0, 1, 2},
			edges: [][3]int{
				{0, 1, 1},
				{1, 2, 1},
			},
			ordering:       "3\n1 1\n2 2\n3 3\n",
			source:         1,
			target:         1,
			expectedPath:   []graph.VertexId{1},
			expectedWeight: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := buildGraph(tc.vertices, tc.edges)
			cch := preprocessAndCustomizeCCH(t, g, tc.ordering)

			path, weight, _, err := cch.Query(tc.source, tc.target)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error for unreachable target, but got nil")
				} else if !errors.Is(err, pathfinding.ErrTargetNotReachable) {
					t.Errorf("Expected error to wrap %v, but got %v", pathfinding.ErrTargetNotReachable, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Query failed: %v", err)
			}

			if !reflect.DeepEqual(path, tc.expectedPath) {
				t.Errorf("Expected path %v, but got %v", tc.expectedPath, path)
			}

			if weight != tc.expectedWeight {
				t.Errorf("Expected weight %f, but got %f", tc.expectedWeight, weight)
			}
		})
	}
}
