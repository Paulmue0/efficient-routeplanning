package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

// Helper function to create a temporary test directory with the specified content.
func setupTestDir(t *testing.T, filename, content string) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "parser-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	tempFile := filepath.Join(tempDir, filename)
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to write temp file: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	return tempDir, cleanup
}

func TestToMetis(t *testing.T) {
	tests := []struct {
		name        string
		graph       *graph.Graph
		expected    string
		expectError bool
	}{
		{
			name: "simple manually created graph",
			graph: func() *graph.Graph {
				g := graph.NewGraph()
				g.AddVertex(graph.Vertex{Id: 0})
				g.AddVertex(graph.Vertex{Id: 1})
				g.AddVertex(graph.Vertex{Id: 2})
				g.AddVertex(graph.Vertex{Id: 3})
				g.AddEdge(0, 1, 1, false, -1)
				g.AddEdge(0, 2, 1, false, -1)
				g.AddEdge(0, 3, 1, false, -1)
				return g
			}(),
			expected: "4 3\n2 3 4\n1\n1\n1\n",
		},
		{
			name: "more complex graph",
			graph: func() *graph.Graph {
				g := graph.NewGraph()
				g.AddVertex(graph.Vertex{Id: 0})
				g.AddVertex(graph.Vertex{Id: 1})
				g.AddVertex(graph.Vertex{Id: 2})
				g.AddEdge(0, 1, 1, false, -1)
				g.AddEdge(1, 2, 1, false, -1)
				return g
			}(),
			expected: "3 2\n2\n1 3\n2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ToMetis(tt.graph, &buf)

			if err != nil && !tt.expectError {
				t.Fatalf("ToMetis returned an unexpected error: %v", err)
			}
			if err == nil && tt.expectError {
				t.Fatalf("ToMetis was expected to return an error, but didn't")
			}

			if actual := buf.String(); actual != tt.expected {
				t.Errorf("ToMetis output mismatch.\nExpected:\n%q\nActual:\n%q", tt.expected, actual)
			}
		})
	}
}

func TestToMetisFromFile(t *testing.T) {
	// Note: This test requires a file to exist at the specified path.
	name := "example.txt"
	dataDir := "../data/RoadNetworks"
	fileSystem := os.DirFS(dataDir)
	network, err := NewNetworkFromFS(fileSystem, name)
	if err != nil {
		t.Fatalf("NewNetworkFromFS returned an error: %v", err)
	}

	expectedMetisOutput := "4 3\n2 3 4\n1\n1\n1\n"

	var buf bytes.Buffer
	err = ToMetis(network.Network, &buf)
	if err != nil {
		t.Fatalf("ToMetis returned an error: %v", err)
	}

	actualMetisOutput := buf.String()
	if actualMetisOutput != expectedMetisOutput {
		t.Errorf("ToMetis output mismatch.\nExpected:\n%q\nActual:\n%q", expectedMetisOutput, actualMetisOutput)
	}
}
