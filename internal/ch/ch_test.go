package ch

import (
	"container/heap"
	"fmt"
	"os"
	"reflect"
	"testing"

	parser "github.com/PaulMue0/efficient-routeplanning/internal/parser"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
	collection "github.com/PaulMue0/efficient-routeplanning/pkg/collection/heap_gen"
)

// --- Test Helper Functions ---

// createGraphFromSlidedeck creates a graph with a known ambiguity
// to ensure tests can handle non-deterministic outcomes.
// The path between nodes 4 and 5 has two shortest paths of equal weight.
func createGraphFromSlidedeck() *graph.Graph {
	g := graph.NewGraph()

	for i := 0; i <= 7; i++ {
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
		{1, 4, 3}, // Original weight, creates ambiguity
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

// getEdge is a test helper to safely access an edge and report a test failure if it doesn't exist.
func getEdge(t *testing.T, g *graph.Graph, from, to graph.VertexId) (graph.Edge, bool) {
	t.Helper()
	if _, ok := g.Edges[from]; !ok {
		return graph.Edge{}, false
	}
	edge, exists := g.Edges[from][to]
	if !exists {
		return graph.Edge{}, false
	}
	return edge, true
}

// assertIsPermutation checks if a slice of VertexIds is a valid permutation
// of the vertices from an original graph.
func assertIsPermutation(t *testing.T, got []graph.VertexId, originalVertices map[graph.VertexId]graph.Vertex) {
	t.Helper()

	if len(got) != len(originalVertices) {
		t.Errorf("Contraction order length is incorrect: got %d, want %d", len(got), len(originalVertices))
		return
	}

	seen := make(map[graph.VertexId]bool)
	for _, vID := range got {
		if _, exists := originalVertices[vID]; !exists {
			t.Errorf("Contraction order contains vertex %d which was not in the original graph", vID)
			return
		}
		if seen[vID] {
			t.Errorf("Contraction order contains duplicate vertex %d", vID)
			return
		}
		seen[vID] = true
	}
}

// --- Unit Tests ---

func TestNumShortcuts(t *testing.T) {
	g := createGraphFromSlidedeck()

	shortcutTests := []struct {
		vId   graph.VertexId
		wants graph.VertexId
	}{
		{vId: 0, wants: 1},
		{vId: 1, wants: 5},
		{vId: 2, wants: 0},
		{vId: 3, wants: 0},
		{vId: 4, wants: 2},
		{vId: 5, wants: 1},
		{vId: 6, wants: 0},
		{vId: 7, wants: 0},
	}

	for _, tt := range shortcutTests {
		t.Run(string(rune(tt.vId)), func(t *testing.T) {
			got := Shortcuts(g, tt.vId, false)
			if int(tt.wants) != got {
				t.Errorf("got %v, want one of %v", got, tt.wants)
			}
		})
	}
}

func TestEdgeDifference(t *testing.T) {
	g := createGraphFromSlidedeck()

	edTests := []struct {
		name  string
		vId   graph.VertexId
		wants graph.VertexId
	}{
		{name: "VertexId:0", vId: 0, wants: -1},
		{name: "VertexId:1", vId: 1, wants: 0},
		{name: "VertexId:2", vId: 2, wants: -2},
		{name: "VertexId:3", vId: 3, wants: -1},
		{name: "VertexId:4", vId: 4, wants: -1},
		{name: "VertexId:5", vId: 5, wants: -1},
		{name: "VertexId:6", vId: 6, wants: -2},
		{name: "VertexId:7", vId: 7, wants: -1},
	}

	for _, tt := range edTests {
		t.Run(tt.name, func(t *testing.T) {
			got := EdgeDifference(g, tt.vId)
			if int(tt.wants) != got {
				t.Errorf("got %v, want one of %v", got, tt.wants)
			}
		})
	}
}

func TestPreprocess(t *testing.T) {
	g := createGraphFromSlidedeck()
	originalVertices := make(map[graph.VertexId]graph.Vertex)
	for id, v := range g.Vertices {
		originalVertices[id] = v
	}

	ch := NewContractionHierarchies()
	ch.Preprocess(g)

	assertIsPermutation(t, ch.ContractionOrder, originalVertices)
}

func TestNewContractionHierarchies(t *testing.T) {
	ch := NewContractionHierarchies()

	if ch.ContractionOrder == nil {
		t.Error("ContractionOrder should be an initialized slice, not nil")
	}
	if len(ch.ContractionOrder) != 0 {
		t.Errorf("ContractionOrder should be empty, got len %d", len(ch.ContractionOrder))
	}
	if ch.Priorities == nil {
		t.Error("Priorities priority queue should not be nil")
	}
	if ch.UpwardsGraph == nil || ch.DownwardsGraph == nil {
		t.Error("UpwardsGraph and DownwardsGraph should not be nil")
	}
	if len(ch.UpwardsGraph.Vertices) != 0 || len(ch.DownwardsGraph.Vertices) != 0 {
		t.Error("UpwardsGraph and DownwardsGraph should be empty")
	}
}

func TestInitializePriority(t *testing.T) {
	g := createGraphFromSlidedeck()
	ch := NewContractionHierarchies()

	ch.InitializePriority(g)

	if ch.Priorities.Len() != len(g.Vertices) {
		t.Errorf("Priority queue should have %d items, but has %d", len(g.Vertices), ch.Priorities.Len())
	}

	// The minimum possible edge difference in the ambiguous graph is -2 (from node 2, and possibly node 6).
	// We test the priority value, not which node comes first.
	item := heap.Pop(ch.Priorities).(*collection.Item[graph.VertexId])
	gotPriority := ch.Priorities.GetPriority(item)
	wantPriority := -2.0

	if gotPriority != wantPriority {
		t.Errorf("Expected highest priority (minimum value) to be %f, but got %f", wantPriority, gotPriority)
	}
}

func TestContract(t *testing.T) {
	g := createGraphFromSlidedeck()
	ch := NewContractionHierarchies()

	// We will contract vertex 0.
	// Neighbors are 1 and 2. The path (1-0-2) has cost 2+1=3.
	// The direct edge (1,2) has cost 4.
	// So, contracting 0 should create a shortcut (1,2) with weight 3.
	contractNodeID := graph.VertexId(0)

	ch.Contract(g, contractNodeID)

	// 1. Check Contraction Order
	if len(ch.ContractionOrder) != 1 || ch.ContractionOrder[0] != contractNodeID {
		t.Errorf("ContractionOrder should be [%d], but got %v", contractNodeID, ch.ContractionOrder)
	}

	// 2. Check if node was removed from the original graph
	if _, exists := g.Vertices[contractNodeID]; exists {
		t.Errorf("Vertex %d should have been removed from the graph", contractNodeID)
	}

	// 3. Check if shortcut was added
	edge, exists := getEdge(t, g, 1, 2)
	if !exists || !edge.IsShortcut {
		t.Fatalf("Expected shortcut edge between 1 and 2, but it was not found")
	}
	if edge.Weight != 3 {
		t.Errorf("Shortcut (1,2) has wrong weight: got %d, want 3", edge.Weight)
	}
	if !edge.IsShortcut {
		t.Error("Edge (1,2) should be marked as a shortcut")
	}
	if edge.Via != contractNodeID {
		t.Errorf("Shortcut (1,2) should be via node %d, but got %d", contractNodeID, edge.Via)
	}
}

func TestInsertInUpwardsOrDownwardsGraph(t *testing.T) {
	t.Run("inserting first node", func(t *testing.T) {
		g := createGraphFromSlidedeck()
		ch := NewContractionHierarchies()
		nodeToInsert := graph.VertexId(2) // Neighbors are 0 and 1

		// Simulate that '2' is the first node being contracted
		ch.ContractionOrder = append(ch.ContractionOrder, nodeToInsert)

		ch.InsertInUpwardsOrDownwardsGraph(g, nodeToInsert)

		// Since no other node is in the order, all edges must go "up" from node 2.
		// UpwardsGraph: 2 -> 0 and 2 -> 1
		// DownwardsGraph: 0 -> 2 and 1 -> 2
		if _, exists := getEdge(t, ch.UpwardsGraph, 2, 0); !exists {
			t.Error("UpwardsGraph missing edge (2 -> 0)")
		}
		if _, exists := getEdge(t, ch.DownwardsGraph, 0, 2); !exists {
			t.Error("DownwardsGraph missing edge (0 -> 2)")
		}
		if _, exists := getEdge(t, ch.UpwardsGraph, 2, 1); !exists {
			t.Error("UpwardsGraph missing edge (2 -> 1)")
		}
		if _, exists := getEdge(t, ch.DownwardsGraph, 1, 2); !exists {
			t.Error("DownwardsGraph missing edge (1 -> 2)")
		}
	})

	t.Run("inserting a subsequent node", func(t *testing.T) {
		g := createGraphFromSlidedeck()
		ch := NewContractionHierarchies()

		// Simulate that '2' was already contracted.
		ch.ContractionOrder = append(ch.ContractionOrder, 2)
		// Now we contract node '0'. Neighbors are 1 and 2.
		nodeToInsert := graph.VertexId(0)
		ch.ContractionOrder = append(ch.ContractionOrder, nodeToInsert)

		ch.InsertInUpwardsOrDownwardsGraph(g, nodeToInsert)

		// Edge (0, 1): 1 is not in the order yet, so it's "above" 0.
		// UpwardsGraph should get (0 -> 1).
		if _, exists := getEdge(t, ch.UpwardsGraph, 0, 1); !exists {
			t.Error("UpwardsGraph missing edge (0 -> 1)")
		}

		// Edge (0, 2): 2 is already in the order, so it's "below" 0.
		// DownwardsGraph should get (0 -> 2).
		if _, exists := getEdge(t, ch.DownwardsGraph, 0, 2); !exists {
			t.Error("DownwardsGraph missing edge (0 -> 2)")
		}
	})
}

func TestQuery(t *testing.T) {
	// Use a copy of the graph for preprocessing because it gets modified.
	gForPreprocess := createGraphFromSlidedeck()
	ch := NewContractionHierarchies()
	ch.Preprocess(gForPreprocess)

	t.Run("Path 4 to 5", func(t *testing.T) {
		source := graph.VertexId(4)
		target := graph.VertexId(5)
		expectedWeight := 8.0
		expectedPath1 := []graph.VertexId{4, 1, 5}
		expectedPath2 := []graph.VertexId{4, 6, 5}

		path, weight, err := ch.Query(source, target)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if weight != expectedWeight {
			t.Errorf("got weight %f, want %f", weight, expectedWeight)
		}

		if !reflect.DeepEqual(path, expectedPath1) && !reflect.DeepEqual(path, expectedPath2) {
			t.Errorf("got path %v, want %v or %v", path, expectedPath1, expectedPath2)
		}
	})

	t.Run("Path 0 to 6", func(t *testing.T) {
		source := graph.VertexId(0)
		target := graph.VertexId(6)
		expectedWeight := 9.0
		expectedPath := []graph.VertexId{0, 1, 5, 6}

		path, weight, err := ch.Query(source, target)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if weight != expectedWeight {
			t.Errorf("got weight %f, want %f", weight, expectedWeight)
		}

		if !reflect.DeepEqual(path, expectedPath) {
			t.Errorf("got path %v, want %v", path, expectedPath)
		}
	})

	t.Run("Path to self", func(t *testing.T) {
		source := graph.VertexId(3)
		target := graph.VertexId(3)
		expectedWeight := 0.0
		expectedPath := []graph.VertexId{3}

		path, weight, err := ch.Query(source, target)

		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if weight != expectedWeight {
			t.Errorf("got weight %f, want %f", weight, expectedWeight)
		}

		if !reflect.DeepEqual(path, expectedPath) {
			t.Errorf("got path %v, want %v", path, expectedPath)
		}
	})
}

func BenchmarkOsm1(b *testing.B) {
	for b.Loop() {
		name := "osm1.txt"
		dataDir := "../data/RoadNetworks"
		fileSystem := os.DirFS(dataDir)
		network, err := parser.NewNetworkFromFS(fileSystem, name)
		if err != nil {
			fmt.Println(err)
		}

		ch := NewContractionHierarchies()
		ch.Preprocess(network.Network)

	}
}