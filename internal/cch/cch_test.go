package cch

import (
	"os"
	"path/filepath"
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

// setupTestFile is a helper function to create a temporary file with specified content
// for testing purposes.
func setupTestFile(t *testing.T, filename, content string) string {
	t.Helper()
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, filename)
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	return tempFile
}

// buildGraph helper: builds a graph with given vertices and edges
func buildGraph(vertices []graph.VertexId, edges [][3]int) *graph.Graph {
	g := graph.NewGraph()
	for _, v := range vertices {
		g.AddVertex(graph.Vertex{Id: v})
	}
	for _, e := range edges {
		u := graph.VertexId(e[0])
		v := graph.VertexId(e[1])
		w := e[2]
		// add both directions
		g.AddEdge(u, v, w, false, -1)
		g.AddEdge(v, u, w, false, -1)
	}
	return g
}

// preprocessCCH helper: preprocesses graph with ordering
func preprocessCCH(t *testing.T, g *graph.Graph, ordering string) *CCH {
	t.Helper()
	orderingFile := setupTestFile(t, "ordering.txt", ordering)
	cch := NewCCH()
	if err := cch.Preprocess(g, orderingFile); err != nil {
		t.Fatalf("CCH.Preprocess failed: %v", err)
	}
	return cch
}

// preprocessAndCustomizeCCH helper: preprocesses and customizes a CCH
func preprocessAndCustomizeCCH(t *testing.T, g *graph.Graph, ordering string) *CCH {
	t.Helper()
	cch := preprocessCCH(t, g, ordering)
	if err := cch.Customize(g); err != nil {
		t.Fatalf("CCH.Customize failed: %v", err)
	}
	return cch
}
