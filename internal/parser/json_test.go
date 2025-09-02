package parser

import (
	"bytes"
	"encoding/json"
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

func TestToJSON(t *testing.T) {
	g := graph.NewGraph()
	g.AddVertex(graph.Vertex{Id: 1, Lat: 3.0, Lon: 4.0})
	g.AddVertex(graph.Vertex{Id: 0, Lat: 1.0, Lon: 2.0})
	g.AddEdge(0, 1, 100, false, -1)

	jsonData, err := ToJSON(g)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// After sorting, nodes should be [0, 1]
	expected := `{"nodes":[{"Id":0,"Lat":1,"Lon":2},{"Id":1,"Lat":3,"Lon":4}],"links":[{"source":0,"target":1,"weight":100,"is_shortcut":false, "via":-1}]}`

	// Normalize JSON strings to avoid issues with whitespace and formatting.
	norm := func(s string) string {
		var out bytes.Buffer // Use bytes.Buffer
		json.Compact(&out, []byte(s))
		return out.String()
	}

	if norm(string(jsonData)) != norm(expected) {
		t.Errorf("Expected JSON:\n%s\nGot:\n%s", norm(expected), norm(string(jsonData)))
	}
}
