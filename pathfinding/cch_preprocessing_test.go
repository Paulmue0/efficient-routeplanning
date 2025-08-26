package pathfinding

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

// setupTestFile is a helper function to create a temporary file with specified content
// for testing purposes.
func setupTestFile(t *testing.T, filename, content string) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "cch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	tempFile := filepath.Join(tempDir, filename)
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to write temp file: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return tempFile
}

func TestCreateContractionOrder(t *testing.T) {
	g := graph.NewGraph()
	g.AddVertex(graph.Vertex{Id: 0})
	g.AddVertex(graph.Vertex{Id: 1})
	g.AddVertex(graph.Vertex{Id: 2})
	g.AddVertex(graph.Vertex{Id: 3})
	g.AddVertex(graph.Vertex{Id: 4})

	tests := []struct {
		name          string
		orderingFile  string
		expectedOrder []graph.VertexId
		expectError   bool
	}{
		{
			name:          "Valid ordering file",
			orderingFile:  "5\n4\n3\n2\n1",
			expectedOrder: []graph.VertexId{0, 1, 2, 3, 4},
			expectError:   false,
		},
		{
			name:          "Different valid ordering",
			orderingFile:  "1\n2\n3\n4\n5",
			expectedOrder: []graph.VertexId{4, 3, 2, 1, 0},
			expectError:   false,
		},
		{
			name:          "File with invalid integer",
			orderingFile:  "5\n4\nabc\n2\n1",
			expectedOrder: nil,
			expectError:   true,
		},
		{
			name:          "Mismatched node count",
			orderingFile:  "5\n4\n3\n2",
			expectedOrder: nil,
			expectError:   true,
		},
		{
			name:          "Unknown METIS ID in file",
			orderingFile:  "5\n4\n10\n2\n1",
			expectedOrder: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := setupTestFile(t, "ordering.txt", tt.orderingFile)

			order, err := createContractionOrder(g, filePath)

			if (err != nil) != tt.expectError {
				t.Fatalf("createContractionOrder() error = %v, expectError %v", err, tt.expectError)
			}

			if !reflect.DeepEqual(order, tt.expectedOrder) {
				t.Errorf("createContractionOrder() got = %v, want %v", order, tt.expectedOrder)
			}
		})
	}
}
