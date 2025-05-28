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

func assertError(t testing.TB, got error, want error) {
	t.Helper()

	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func assertInt(t testing.TB, got int, want int) {
	if got != want {
		t.Errorf("expected %d got %d", want, got)
	}
}

func assertStrings(t testing.TB, got string, want string) {
	if got != want {
		t.Errorf("expected %s got %s", want, got)
	}
}

func assertSlice[T Vertex | Edge](t testing.TB, got []T, want []T) {
	if !slices.Equal(got, want) {
		t.Errorf("expected %q got %q", want, got)
	}
}
