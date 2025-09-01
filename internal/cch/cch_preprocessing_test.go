package cch

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

// setupTestFile is a helper function to create a temporary file with specified content
// for testing purposes.
func setupTestFile(t *testing.T, filename, content string) string {
	t.Helper()

	// Create a temporary directory for the test file
	tempDir := t.TempDir()

	tempFile := filepath.Join(tempDir, filename)
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	return tempFile
}

func TestInitializeContraction(t *testing.T) {
	// 1. Setup Graph
	g := graph.NewGraph()
	g.AddVertex(graph.Vertex{Id: 0})
	g.AddVertex(graph.Vertex{Id: 1})
	g.AddVertex(graph.Vertex{Id: 2})

	// Mapping rule: Graph ID = METIS ID - 1
	//   METIS ID 1 -> graph ID 0
	//   METIS ID 2 -> graph ID 1
	//   METIS ID 3 -> graph ID 2

	// 2. Setup ordering file
	// Format: "rank METIS_ID"
	orderingContent := `3
1 3
2 1
3 2
`
	// Which means ranks:
	//   rank 1 -> METIS ID 3 -> Graph ID 2
	//   rank 2 -> METIS ID 1 -> Graph ID 0
	//   rank 3 -> METIS ID 2 -> Graph ID 1
	orderingFile := setupTestFile(t, "ordering.txt", orderingContent)

	// 3. Initialize CCH
	cch := NewCCH()
	err := cch.initializeContraction(g, orderingFile)
	if err != nil {
		t.Fatalf("initializeContraction failed: %v", err)
	}

	// 4. Assertions
	expectedOrder := []graph.VertexId{2, 0, 1}
	if !reflect.DeepEqual(cch.ContractionOrder, expectedOrder) {
		t.Errorf("Expected ContractionOrder %v, got %v", expectedOrder, cch.ContractionOrder)
	}

	expectedMap := map[graph.VertexId]int{
		2: 0,
		0: 1,
		1: 2,
	}
	if !reflect.DeepEqual(cch.ContractionMap, expectedMap) {
		t.Errorf("Expected ContractionMap %v, got %v", expectedMap, cch.ContractionMap)
	}
}

func TestInitializeContraction_Errors(t *testing.T) {
	g := graph.NewGraph()
	g.AddVertex(graph.Vertex{Id: 0})
	g.AddVertex(graph.Vertex{Id: 1})

	tests := []struct {
		name          string
		fileContent   string
		isFileMissing bool
		expectedError string
	}{
		{
			name: "mismatched node count",
			fileContent: `1 1
`,
			expectedError: "mismatch in node count",
		},
		{
			name: "invalid integer",
			fileContent: `1 1
invalid
`,
			expectedError: "mismatch in node count: graph has 2 nodes, but ordering file has 1 entries",
		},
		{
			name: "metis id not found",
			fileContent: `1 1
2 3
`, // METIS ID 3 -> Graph ID 2, which does not exist
			expectedError: "not found",
		},
		{
			name:          "file not found",
			isFileMissing: true,
			expectedError: "failed to open ordering file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var orderingFile string
			if tc.isFileMissing {
				orderingFile = filepath.Join(t.TempDir(), "non_existent", "ordering.txt")
			} else {
				orderingFile = setupTestFile(t, "ordering.txt", tc.fileContent)
			}

			cch := NewCCH()
			err := cch.initializeContraction(g, orderingFile)

			if err == nil {
				t.Fatalf("Expected an error, but got nil")
			}
			if !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("Expected error containing '%s', got '%v'", tc.expectedError, err)
			}
		})
	}
}

func TestCCHPreprocess(t *testing.T) {
	// 1. Setup Graph
	// Graph:
	// 0 -> 1 (10)
	// 0 -> 2 (10)
	// 1 -> 3 (10)
	// 2 -> 3 (10)
	g := graph.NewGraph()
	g.AddVertex(graph.Vertex{Id: 0})
	g.AddVertex(graph.Vertex{Id: 1})
	g.AddVertex(graph.Vertex{Id: 2})
	g.AddVertex(graph.Vertex{Id: 3})
	g.AddEdge(0, 1, 10, false, -1)
	g.AddEdge(1, 0, 10, false, -1)
	g.AddEdge(0, 2, 10, false, -1)
	g.AddEdge(2, 0, 10, false, -1)
	g.AddEdge(1, 3, 10, false, -1)
	g.AddEdge(3, 1, 10, false, -1)
	g.AddEdge(2, 3, 10, false, -1)
	g.AddEdge(3, 2, 10, false, -1)

	// 2. Setup ordering file
	// METIS IDs: graph ID + 1
	// Graph IDs {0,1,2,3} -> METIS {1,2,3,4}
	// Order: rank1->1, rank2->4, rank3->2, rank4->3
	orderingContent := `4
1 1
2 4
3 2
4 3
`
	orderingFile := setupTestFile(t, "ordering.txt", orderingContent)

	// 3. Preprocess
	cch := NewCCH()
	err := cch.Preprocess(g, orderingFile)
	if err != nil {
		t.Fatalf("Preprocess failed: %v", err)
	}

	// 4. Assertions
	expectedOrder := []graph.VertexId{0, 3, 1, 2}
	if !reflect.DeepEqual(cch.ContractionOrder, expectedOrder) {
		t.Errorf("Expected ContractionOrder %v, got %v", expectedOrder, cch.ContractionOrder)
	}
	expectedMap := map[graph.VertexId]int{0: 0, 3: 1, 1: 2, 2: 3}
	if !reflect.DeepEqual(cch.ContractionMap, expectedMap) {
		t.Errorf("Expected ContractionMap %v, got %v", expectedMap, cch.ContractionMap)
	}

	// Check Upwards and Downwards edges
	// Ranks: 0(0), 1(2), 2(3), 3(1)
	// Edge (0,1): rank0 < rank1 => upwards 0->1, downwards 1->0
	// Edge (0,2): rank0 < rank2 => upwards 0->2, downwards 2->0
	// Edge (1,3): rank1 > rank3 => downwards 1->3, upwards 3->1
	// Edge (2,3): rank2 > rank3 => downwards 2->3, upwards 3->2

	upEdges := map[graph.VertexId][]graph.VertexId{
		0: {1, 2},
		3: {1, 2},
	}
	downEdges := map[graph.VertexId][]graph.VertexId{
		1: {0, 3},
		2: {0, 3},
	}

	for u, vs := range upEdges {
		for _, v := range vs {
			if exists, _ := cch.UpwardsGraph.Adjacent(u, v); !exists {
				t.Errorf("Expected edge %d->%d in UpwardsGraph", u, v)
			}
		}
	}
	for u, vs := range downEdges {
		for _, v := range vs {
			if exists, _ := cch.DownwardsGraph.Adjacent(u, v); !exists {
				t.Errorf("Expected edge %d->%d in DownwardsGraph", u, v)
			}
		}
	}

	// Check shortcut
	// Contraction order: [0,3,1,2]
	// Node 0 contracted first. Higher-ranked neighbors: 1,2
	// Path 1-0-2 => shortcut between 1 and 2
	exists, err := cch.UpwardsGraph.Adjacent(1, 2)
	if err != nil {
		t.Fatalf("Error checking adjacency for shortcut: %v", err)
	}
	if !exists {
		t.Fatal("Expected shortcut 1->2 in UpwardsGraph after contraction")
	}
	edge, ok := cch.UpwardsGraph.Edges[1][2]
	if !ok {
		t.Fatal("Edge 1->2 not found in UpwardsGraph")
	}
	if !edge.IsShortcut {
		t.Error("Edge 1->2 should be marked as a shortcut")
	}
	if edge.Via != 0 {
		t.Errorf("Expected shortcut 1->2 to be via node 0, but got %d", edge.Via)
	}
}
