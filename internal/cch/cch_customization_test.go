package cch

import (
	"math"
	"testing"

	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
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
			vertices: []graph.VertexId{0, 1, 2},
			edges:    [][3]int{{0, 1, 10}, {1, 2, 5}},
			// Order 0->1->2 ensures all original edges are upward edges.
			// Format is "rank ID", both 1-based.
			ordering: "3\n1 1\n2 2\n3 3\n",
			check: func(t *testing.T, cch *CCH) {
				assertEdgeWeight(t, cch, 0, 1, 10)
				assertEdgeWeight(t, cch, 1, 2, 5)
			},
		},
		{
			name:     "Shortcuts created in Preprocess are set to infinity",
			vertices: []graph.VertexId{0, 1, 2, 3},
			edges: [][3]int{
				{0, 1, 1},
				{0, 2, 1},
				{0, 3, 10},
				{1, 2, 5}, // direct edge is worse than shortcut 1->0->2 (cost 2)
			},
			// Contract 0 first to create shortcuts between its neighbors (1,2,3).
			// Format is "rank ID", both 1-based.
			ordering: "4\n1 1\n2 2\n3 3\n4 4\n",
			check: func(t *testing.T, cch *CCH) {
				assertEdgeWeight(t, cch, 0, 2, 1)
				assertEdgeWeight(t, cch, 1, 2, 5)
				assertEdgeWeight(t, cch, 0, 1, 1)
				// This shortcut (1->3 via 0) is created during preprocess.
				// Respecting() resets its weight to infinity before customization.
				assertShortcut(t, cch, 1, 3, 0, int(math.Inf(1)))
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
			vertices: []graph.VertexId{0, 1, 2},
			edges: [][3]int{
				{0, 1, 10},
				{0, 2, 1},
				{1, 2, 1}, // Path 1->0->2 has cost 11, which is worse
			},
			// Contract 0 first, so we check for a shortcut between 1 and 2.
			// Format is "rank ID", both 1-based.
			ordering: "3\n1 1\n2 2\n3 3\n",
			check: func(t *testing.T, cch *CCH) {
				assertEdgeWeight(t, cch, 1, 2, 1)
			},
		},
		{
			name:     "shortcut improves edge weight",
			vertices: []graph.VertexId{0, 1, 2},
			edges: [][3]int{
				{0, 1, 1},
				{0, 2, 1},
				{1, 2, 5}, // direct edge (cost 5) is worse than shortcut 1->0->2 (cost 2)
			},
			// Contract 0 first to find the better shortcut path between 1 and 2.
			// Format is "rank ID", both 1-based.
			ordering: "3\n1 1\n2 2\n3 3\n",
			check: func(t *testing.T, cch *CCH) {
				assertShortcut(t, cch, 1, 2, 0, 2) // 1->2 is now a shortcut via 0
			},
		},
		{
			name:     "shortcut creates missing upward edge",
			vertices: []graph.VertexId{0, 1, 2},
			edges: [][3]int{
				{0, 1, 1},
				{0, 2, 1},
				// no direct 1->2 edge
			},
			// Contract 0 first to create a shortcut where no edge existed.
			// Format is "rank ID", both 1-based.
			ordering: "3\n1 1\n2 2\n3 3\n",
			check: func(t *testing.T, cch *CCH) {
				assertEdgeWeight(t, cch, 1, 2, 2) // added via shortcut 1->0->2
			},
		},
		{
			name:     "disconnected nodes remain disconnected",
			vertices: []graph.VertexId{0, 1, 2, 3},
			edges: [][3]int{
				{0, 1, 1}, // Component 1
				{2, 3, 1}, // Component 2
			},
			// Any valid ordering works here.
			// Format is "rank ID", both 1-based.
			ordering: "4\n1 1\n2 2\n3 3\n4 4\n",
			check: func(t *testing.T, cch *CCH) {
				if _, ok := cch.UpwardsGraph.Edges[1][2]; ok {
					t.Errorf("Unexpected edge 1->2 found in disconnected graph")
				}
				if _, ok := cch.UpwardsGraph.Edges[0][3]; ok {
					t.Errorf("Unexpected edge 0->3 found in disconnected graph")
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
