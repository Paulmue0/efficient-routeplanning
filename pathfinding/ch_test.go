package pathfinding

import (
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

func createGraphFromSlidedeck() *graph.Graph {
	g := graph.NewGraph()

	for i := 0; i <= 8; i++ {
		g.AddVertex(graph.Vertex{Id: graph.VertexId(i)})
	}

	edges := []struct {
		from, to graph.VertexId
		weight   int
	}{
		{0, 1, 2},
		{0, 2, 1},
		{1, 3, 10},
		{1, 4, 3},
		{1, 5, 5},
		{4, 6, 6},
		{4, 7, 9},
		{5, 6, 2},
	}

	for _, e := range edges {
		g.AddEdge(e.from, e.to, e.weight)
		g.AddEdge(e.to, e.from, e.weight)
	}

	return g
}

func TestNumShortcuts(t *testing.T) {
	g := createGraphFromSlidedeck()
	got := 1
	want := NumShortcuts(g, 0)

	if got != want {
		t.Errorf("wanted %v, got %v", got, want)
	}
}
