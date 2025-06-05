package graph

import (
	"slices"
	"testing"
)

func TestSearch(t *testing.T) {
	graph := NewGraph()
	vertex := Vertex{0}
	graph.AddVertex(vertex)

	t.Run("known vertex", func(t *testing.T) {
		got, _ := graph.Search(vertex)
		want := []Edge{}

		if !slices.Equal(got, want) {
			t.Errorf("expected %v go %v", want, got)
		}
	})

	t.Run("unknown vertex", func(t *testing.T) {
		_, err := graph.Search(Vertex{1})
		want := ErrVertexNotFound

		assertError(t, err, want)
		assertStrings(t, err.Error(), want.Error())
	})
}

func TestAddVertex(t *testing.T) {
	t.Run("adding a new vertex", func(t *testing.T) {
		g := NewGraph()
		g.AddVertex(Vertex{})

		got := len(g.AdjacencyList)
		want := 1

		assertInt(t, got, want)
	})
	t.Run("adding existing vertex", func(t *testing.T) {
		g := NewGraph()
		g.AddVertex(Vertex{1})

		err := g.AddVertex(Vertex{1})
		want := ErrVertexAlreadyExists

		assertError(t, err, want)
	})
}

func TestRemoveVertex(t *testing.T) {
	t.Run("remove existing vertex", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{1}
		vertex2 := Vertex{2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)
		g.RemoveVertex(vertex)

		got := len(g.AdjacencyList)
		want := 1

		assertInt(t, got, want)
	})
	t.Run("not existing vertex", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{1}
		vertex2 := Vertex{2}
		vertex3 := Vertex{3}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)
		err := g.RemoveVertex(vertex3)

		want := ErrVertexNotFound

		assertError(t, err, want)
	})
	t.Run("remove incoming edges", func(t *testing.T) {
		g := NewGraph()
		v1 := Vertex{1}
		v2 := Vertex{2}
		v3 := Vertex{3}

		g.AddVertex(v1)
		g.AddVertex(v2)
		g.AddVertex(v3)

		g.AddEdge(v2, Edge{Target: v1, Weight: 5})
		g.AddEdge(v3, Edge{Target: v1, Weight: 10})

		err := g.RemoveVertex(v1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, exists := g.AdjacencyList[v1]; exists {
			t.Errorf("expected vertex %v to be removed", v1)
		}

		for _, v := range []Vertex{v2, v3} {
			for _, edge := range g.AdjacencyList[v] {
				if edge.Target == v1 {
					t.Errorf("expected no incoming edge to %v from %v", v1, v)
				}
			}
		}
	})
	t.Run("idempotency", func(t *testing.T) {
		g := NewGraph()
		v := Vertex{1}
		g.AddVertex(v)

		_ = g.RemoveVertex(v)
		err := g.RemoveVertex(v)

		assertError(t, err, ErrVertexNotFound)
	})
}

func TestAddEdge(t *testing.T) {
	t.Run("adding a new edge", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{1}
		vertex2 := Vertex{2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)
		edge := Edge{vertex2, 1}

		g.AddEdge(vertex, edge)
		got := len(g.AdjacencyList[vertex])
		want := 1

		assertInt(t, got, want)
	})
	t.Run("adding existing edge", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{1}
		vertex2 := Vertex{2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)
		edge := Edge{vertex2, 1}

		g.AddEdge(vertex, edge)
		err := g.AddEdge(vertex, edge)
		want := ErrEdgeAlreadyExists

		assertError(t, err, want)
	})
	t.Run("vertex does not exist", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{1}
		vertex2 := Vertex{2}
		g.AddVertex(vertex2)
		edge := Edge{vertex2, 1}

		err := g.AddEdge(vertex, edge)
		want := ErrVertexNotFound

		assertError(t, err, want)
	})
}

func TestRemoveEdge(t *testing.T) {
	t.Run("existing edge", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{1}
		vertex2 := Vertex{2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)
		edge := Edge{vertex2, 1}

		g.AddEdge(vertex, edge)
		g.RemoveEdge(vertex, edge)
		got := len(g.AdjacencyList[vertex])
		want := 0

		assertInt(t, got, want)
	})
	t.Run("not existing edge", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{1}
		vertex2 := Vertex{2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)
		edge := Edge{vertex2, 1}

		err := g.RemoveEdge(vertex, edge)
		want := ErrEdgeNotFound

		assertError(t, err, want)
	})
}

func TestNeighbors(t *testing.T) {
	t.Run("existing neighbor", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{1}
		vertex2 := Vertex{2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)
		edge := Edge{vertex2, 1}
		g.AddEdge(vertex, edge)

		got, _ := g.Neighbors(vertex)
		want := []Vertex{vertex2}

		assertSlice(t, got, want)
	})
	t.Run("no neighbor", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{1}
		g.AddVertex(vertex)

		got, _ := g.Neighbors(vertex)
		want := make([]Vertex, 0)

		assertSlice(t, got, want)
	})
	t.Run("vertex does not exist", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{1}

		_, err := g.Neighbors(vertex)
		want := ErrVertexNotFound

		assertError(t, err, want)
	})
}

func TestAdjacent(t *testing.T) {
	g := NewGraph()
	v1 := Vertex{1}
	v2 := Vertex{2}
	g.AddVertex(v1)
	g.AddVertex(v2)
	g.AddEdge(v1, Edge{Target: v2, Weight: 1})

	t.Run("is adjacent", func(t *testing.T) {
		ok, err := g.Adjacent(v1, v2)
		assertError(t, err, nil)
		if !ok {
			t.Errorf("expected vertices to be adjacent")
		}
	})

	t.Run("not adjacent", func(t *testing.T) {
		v3 := Vertex{3}
		g.AddVertex(v3)
		ok, err := g.Adjacent(v2, v3)
		assertError(t, err, nil)
		if ok {
			t.Errorf("expected vertices to not be adjacent")
		}
	})

	t.Run("unknown vertex", func(t *testing.T) {
		_, err := g.Adjacent(Vertex{99}, v1)
		assertError(t, err, ErrVertexNotFound)
	})
}

func assertError(t testing.TB, got error, want error) {
	t.Helper()

	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func assertInt(t testing.TB, got int, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected %d got %d", want, got)
	}
}

func assertStrings(t testing.TB, got string, want string) {
	t.Helper()
	if got != want {
		t.Errorf("expected %s got %s", want, got)
	}
}

func assertSlice[T Vertex | Edge](t testing.TB, got []T, want []T) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("expected %q got %q", want, got)
	}
}
