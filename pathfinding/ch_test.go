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
		{1, 2, 4},
		{1, 3, 10},
		{1, 4, 3},
		{1, 5, 5},
		{4, 6, 6},
		{4, 7, 9},
		{5, 6, 2},
	}

	for _, e := range edges {
		g.AddEdge(e.from, e.to, e.weight, false, -1)
		g.AddEdge(e.to, e.from, e.weight, false, -1)
	}

	return g
}

func TestNumShortcuts(t *testing.T) {
	g := createGraphFromSlidedeck()

	shortcutTests := map[graph.VertexId]int{
		0: 1,
		1: 6,
		2: 0,
		3: 0,
		4: 2,
		5: 1,
		6: 0,
		7: 0,
	}

	for id, want := range shortcutTests {
		t.Run(string(rune(id)), func(t *testing.T) {
			got := Shortcuts(g, id, false)
			if got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

func TestEdgeDifference(t *testing.T) {
	g := createGraphFromSlidedeck()

	edTests := []struct {
		name string
		vId  graph.VertexId
		want int
	}{
		{name: "VertexId:0", vId: 0, want: -1},
		{name: "VertexId:1", vId: 1, want: 1},
		{name: "VertexId:2", vId: 2, want: -2},
		{name: "VertexId:3", vId: 3, want: -1},
		{name: "VertexId:4", vId: 4, want: -1},
		{name: "VertexId:5", vId: 5, want: -1},
		{name: "VertexId:6", vId: 6, want: -1},
		{name: "VertexId:7", vId: 7, want: -1},
	}

	for _, tt := range edTests {
		t.Run(tt.name, func(t *testing.T) {
			got := EdgeDifference(g, tt.vId)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
