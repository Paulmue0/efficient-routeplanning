package pathfinding

import (
	"math"
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

// helper: builds a graph with given vertices and edges
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

// helper: preprocesses graph with ordering
func preprocessCCH(t *testing.T, g *graph.Graph, ordering string) *CCH {
	t.Helper()
	orderingFile := setupTestFile(t, "ordering.txt", ordering)
	cch := NewCCH()
	if err := cch.Preprocess(g, orderingFile); err != nil {
		t.Fatalf("CCH.Preprocess failed: %v", err)
	}
	return cch
}

// helper: assert edge weight
func assertEdgeWeight(t *testing.T, cch *CCH, u, v graph.VertexId, expected int) {
	t.Helper()
	edge, ok := cch.UpwardsGraph.Edges[u][v]
	if !ok {
		t.Fatalf("Expected edge %d->%d but not found", u, v)
	}
	if edge.Weight != expected {
		t.Errorf("Expected weight of %d->%d = %d, got %d", u, v, expected, edge.Weight)
	}
}

// helper: assert edge is a shortcut with given via node
func assertShortcut(t *testing.T, cch *CCH, u, v, via graph.VertexId, expectedWeight int) {
	t.Helper()
	edge, ok := cch.UpwardsGraph.Edges[u][v]
	if !ok {
		t.Fatalf("Expected edge %d->%d but not found", u, v)
	}
	if !edge.IsShortcut {
		t.Errorf("Expected %d->%d to be a shortcut, but it was not", u, v)
	}
	if edge.Via != via {
		t.Errorf("Expected shortcut %d->%d via %d, got via %d", u, v, via, edge.Via)
	}
	if edge.Weight != expectedWeight {
		t.Errorf("Expected shortcut %d->%d weight %d, got %d", u, v, expectedWeight, edge.Weight)
	}
}

func TestRespecting(t *testing.T) {
	tests := []struct {
		name     string
		vertices []graph.VertexId
		edges    [][3]int
		ordering string
		check    func(t *testing.T, cch *CCH)
	}{
		{
			name:     "Base graph with no shortcuts",
			vertices: []graph.VertexId{1, 2, 3},
			edges:    [][3]int{{1, 2, 10}, {2, 3, 5}},
			ordering: "3\n2\n1\n",
			check: func(t *testing.T, cch *CCH) {
				assertEdgeWeight(t, cch, 1, 2, 10)
				assertEdgeWeight(t, cch, 2, 3, 5)
			},
		},
		{
			name:     "Shortcuts created in Preprocess are set to infinity",
			vertices: []graph.VertexId{1, 2, 3, 4},
			edges: [][3]int{
				{1, 2, 1},
				{1, 3, 1},
				{1, 4, 10},
				{2, 3, 5}, // direct worse than shortcut
			},
			ordering: "4\n3\n2\n1\n",
			check: func(t *testing.T, cch *CCH) {
				assertEdgeWeight(t, cch, 1, 3, 1)
				assertEdgeWeight(t, cch, 2, 3, 5)
				assertEdgeWeight(t, cch, 1, 2, 1)
				assertShortcut(t, cch, 2, 4, 1, int(math.Inf(1)))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalGraph := buildGraph(tc.vertices, tc.edges)
			cch := preprocessCCH(t, originalGraph, tc.ordering)

			if err := cch.Respecting(originalGraph); err != nil {
				t.Fatalf("Respecting failed: %v", err)
			}

			tc.check(t, cch)
		})
	}
}

func TestBasicCustomization(t *testing.T) {
	tests := []struct {
		name     string
		vertices []graph.VertexId
		edges    [][3]int // [u, v, weight]
		ordering string
		check    func(t *testing.T, cch *CCH)
	}{
		{
			name:     "existing edge should not be overwritten by worse shortcut",
			vertices: []graph.VertexId{1, 2, 3},
			edges: [][3]int{
				{1, 2, 10},
				{1, 3, 1},
				{2, 3, 1},
			},
			ordering: "3\n2\n1\n",
			check: func(t *testing.T, cch *CCH) {
				assertEdgeWeight(t, cch, 2, 3, 1)
			},
		},
		{
			name:     "shortcut improves edge weight",
			vertices: []graph.VertexId{1, 2, 3},
			edges: [][3]int{
				{1, 2, 1},
				{1, 3, 1},
				{2, 3, 5}, // direct worse than shortcut
			},
			ordering: "3\n2\n1\n",
			check: func(t *testing.T, cch *CCH) {
				assertShortcut(t, cch, 2, 3, 1, 2) // 2->3 is now a shortcut via 1
			},
		},
		{
			name:     "shortcut creates missing upward edge",
			vertices: []graph.VertexId{1, 2, 3},
			edges: [][3]int{
				{1, 2, 1},
				{1, 3, 1},
				// no direct 2->3
			},
			ordering: "3\n2\n1\n",
			check: func(t *testing.T, cch *CCH) {
				assertEdgeWeight(t, cch, 2, 3, 2) // added via shortcut
			},
		},
		{
			name:     "disconnected nodes remain disconnected",
			vertices: []graph.VertexId{1, 2, 3, 4},
			edges: [][3]int{
				{1, 2, 1},
				{3, 4, 1},
			},
			ordering: "4\n3\n2\n1\n",
			check: func(t *testing.T, cch *CCH) {
				if _, ok := cch.UpwardsGraph.Edges[2][3]; ok {
					t.Errorf("Unexpected edge 2->3 found in disconnected graph")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := buildGraph(tc.vertices, tc.edges)
			cch := preprocessCCH(t, g, tc.ordering)

			if err := cch.Customize(g); err != nil {
				t.Fatalf("Customize failed: %v", err)
			}

			tc.check(t, cch)
		})
	}
}
