package pathfinding

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
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
	// Add vertices out of order to test sorting
	g.AddVertex(graph.Vertex{Id: 30})
	g.AddVertex(graph.Vertex{Id: 10})
	g.AddVertex(graph.Vertex{Id: 20})

	// The mapping from metis is based on sorted vertex IDs.
	// So, 10 -> 1, 20 -> 2, 30 -> 3

	// 2. Setup ordering file
	// METIS IDs in the file: 3, 1, 2
	// This corresponds to graph IDs: 30, 10, 20
	orderingContent := `3
1
2
`
	orderingFile := setupTestFile(t, "ordering.txt", orderingContent)

	// 3. Initialize CCH and call initializeContraction
	cch := NewCCH()
	err := cch.initializeContraction(g, orderingFile)
	if err != nil {
		t.Fatalf("initializeContraction failed: %v", err)
	}

	// 4. Assertions
	// The node ordering is [30, 10, 20]
	// The contraction order is the reverse: [20, 10, 30]
	expectedOrder := []graph.VertexId{20, 10, 30}
	if !reflect.DeepEqual(cch.ContractionOrder, expectedOrder) {
		t.Errorf("Expected ContractionOrder %v, got %v", expectedOrder, cch.ContractionOrder)
	}

	// The contraction map gives the rank: 20 -> 0, 10 -> 1, 30 -> 2
	expectedMap := map[graph.VertexId]int{
		20: 0,
		10: 1,
		30: 2,
	}
	if !reflect.DeepEqual(cch.ContractionMap, expectedMap) {
		t.Errorf("Expected ContractionMap %v, got %v", expectedMap, cch.ContractionMap)
	}
}

func TestInitializeContraction_Errors(t *testing.T) {
	g := graph.NewGraph()
	g.AddVertex(graph.Vertex{Id: 1})
	g.AddVertex(graph.Vertex{Id: 2})

	tests := []struct {
		name          string
		fileContent   string
		isFileMissing bool
		expectedError string
	}{
		{
			name: "mismatched node count",
			fileContent: `1
`,
			expectedError: "mismatch in node count",
		},
		{
			name: "invalid integer",
			fileContent: `1
invalid
`,
			expectedError: "invalid integer found in ordering file",
		},
		{
			name: "metis id not found",
			fileContent: `1
3
`, // METIS ID 3 does not exist
			expectedError: "METIS ID 3 from file not found in graph mapping",
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
				// Create a path in a non-existent directory to ensure it fails
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

// TestPreprocess tests the full CCH preprocessing pipeline.
func TestCCHPreprocess(t *testing.T) {
	// 1. Setup Graph
	// This is a simple graph for testing.
	// 1 -> 2 (w:10)
	// 1 -> 3 (w:10)
	// 2 -> 4 (w:10)
	// 3 -> 4 (w:10)
	g := graph.NewGraph()
	g.AddVertex(graph.Vertex{Id: 1})
	g.AddVertex(graph.Vertex{Id: 2})
	g.AddVertex(graph.Vertex{Id: 3})
	g.AddVertex(graph.Vertex{Id: 4})
	g.AddEdge(1, 2, 10, false, -1)
	g.AddEdge(2, 1, 10, false, -1)
	g.AddEdge(1, 3, 10, false, -1)
	g.AddEdge(3, 1, 10, false, -1)
	g.AddEdge(2, 4, 10, false, -1)
	g.AddEdge(4, 2, 10, false, -1)
	g.AddEdge(3, 4, 10, false, -1)
	g.AddEdge(4, 3, 10, false, -1)

	// 2. Setup ordering file
	// Sorted IDs: 1, 2, 3, 4. Metis IDs: 1->1, 2->2, 3->3, 4->4
	// Let's say the ordering from METIS is 1, 4, 2, 3
	// This means node ordering is [1, 4, 2, 3]
	// Contraction order is reverse: [3, 2, 4, 1]
	// Contraction map: 3->0, 2->1, 4->2, 1->3
	orderingContent := `1
4
2
3
`
	orderingFile := setupTestFile(t, "ordering.txt", orderingContent)

	// 3. Preprocess
	cch := NewCCH()
	err := cch.Preprocess(g, orderingFile)
	if err != nil {
		t.Fatalf("Preprocess failed: %v", err)
	}

	// 4. Assertions
	// Check contraction order and map
	expectedOrder := []graph.VertexId{3, 2, 4, 1}
	if !reflect.DeepEqual(cch.ContractionOrder, expectedOrder) {
		t.Errorf("Expected ContractionOrder %v, got %v", expectedOrder, cch.ContractionOrder)
	}
	expectedMap := map[graph.VertexId]int{3: 0, 2: 1, 4: 2, 1: 3}
	if !reflect.DeepEqual(cch.ContractionMap, expectedMap) {
		t.Errorf("Expected ContractionMap %v, got %v", expectedMap, cch.ContractionMap)
	}

	// Check Upwards and Downwards graphs for initial edges
	// Ranks: 1(3), 2(1), 3(0), 4(2)
	// Edge (1,2): rank(1)=3 > rank(2)=1. Downwards: 1->2, Upwards: 2->1
	// Edge (1,3): rank(1)=3 > rank(3)=0. Downwards: 1->3, Upwards: 3->1
	// Edge (2,4): rank(2)=1 < rank(4)=2. Upwards: 2->4, Downwards: 4->2
	// Edge (3,4): rank(3)=0 < rank(4)=2. Upwards: 3->4, Downwards: 4->3

	upEdges := map[graph.VertexId][]graph.VertexId{
		2: {1, 4},
		3: {1, 4},
	}
	downEdges := map[graph.VertexId][]graph.VertexId{
		1: {2, 3},
		4: {2, 3},
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

	// Check shortcuts
	// The contraction order is [3, 2, 4, 1].
	// Node 3 is contracted first. Its higher-ranked neighbors are 1 and 4,
	// connected by the path 1-3-4. A shortcut (1,4) must be added.
	// Since rank(4) < rank(1), the directed shortcut is 4 -> 1.
	exists, err := cch.UpwardsGraph.Adjacent(4, 1)
	if err != nil {
		t.Fatalf("Error checking adjacency for shortcut: %v", err)
	}
	if !exists {
		t.Fatal("Expected shortcut 4->1 in UpwardsGraph after contraction")
	}
	edge, ok := cch.UpwardsGraph.Edges[4][1]
	if !ok {
		t.Fatal("Edge 4->1 not found in UpwardsGraph")
	}
	if !edge.IsShortcut {
		t.Error("Edge 4->1 should be marked as a shortcut")
	}
	if edge.Via != 3 {
		t.Errorf("Expected shortcut 4->1 to be via node 3, but got %d", edge.Via)
	}
}
