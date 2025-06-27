package graph

import (
	"slices"
	"testing"
)

// func TestSearch(t *testing.T) {
// 	graph := NewGraph()
// 	vertex := Vertex{Id: 0}
// 	graph.AddVertex(vertex)
//
// 	t.Run("known vertex", func(t *testing.T) {
// 		got, _ := graph.Vertex(vertex.Id)
// 		want := vertex
//
// 		assertSlice
// 	})
//
// 	t.Run("unknown vertex", func(t *testing.T) {
// 		_, err := graph.Search(1)
// 		want := ErrVertexNotFound
//
// 		assertError(t, err, want)
// 		assertStrings(t, err.Error(), want.Error())
// 	})
// }

func TestAddVertex(t *testing.T) {
	t.Run("adding a new vertex", func(t *testing.T) {
		g := NewGraph()
		g.AddVertex(Vertex{})

		got := len(g.Vertices)
		want := 1

		assertInt(t, got, want)
	})
	t.Run("adding existing vertex", func(t *testing.T) {
		g := NewGraph()
		g.AddVertex(Vertex{Id: 1})

		err := g.AddVertex(Vertex{Id: 1})
		want := ErrVertexAlreadyExists

		assertError(t, err, want)
	})
}

func TestRemoveVertex(t *testing.T) {
	t.Run("remove existing vertex", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{Id: 1}
		vertex2 := Vertex{Id: 2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)
		g.RemoveVertex(vertex.Id)

		got := len(g.Vertices)
		want := 1

		assertInt(t, got, want)
	})
	t.Run("not existing vertex", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{Id: 1}
		vertex2 := Vertex{Id: 2}
		vertex3 := Vertex{Id: 3}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)
		err := g.RemoveVertex(vertex3.Id)

		want := ErrVertexNotFound

		assertError(t, err, want)
	})
	t.Run("remove with edges", func(t *testing.T) {
		g := NewGraph()
		v1 := Vertex{Id: 1}
		v2 := Vertex{Id: 2}

		g.AddVertex(v1)
		g.AddVertex(v2)

		g.AddEdge(v2.Id, v1.Id, 5)

		err := g.RemoveVertex(v1.Id)
		assertError(t, err, ErrVertexHasEdges)
	})

	t.Run("idempotency", func(t *testing.T) {
		g := NewGraph()
		v := Vertex{Id: 1}
		g.AddVertex(v)

		_ = g.RemoveVertex(v.Id)
		err := g.RemoveVertex(v.Id)

		assertError(t, err, ErrVertexNotFound)
	})
}

func TestAddEdge(t *testing.T) {
	t.Run("adding a new edge", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{Id: 1}
		vertex2 := Vertex{Id: 2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)

		g.AddEdge(vertex.Id, vertex2.Id, 1)
		got := len(g.Edges[vertex.Id])
		want := 1

		assertInt(t, got, want)
	})
	t.Run("adding existing edge", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{Id: 1}
		vertex2 := Vertex{Id: 2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)

		g.AddEdge(vertex.Id, vertex2.Id, 1)
		err := g.AddEdge(vertex.Id, vertex2.Id, 1)
		want := ErrEdgeAlreadyExists

		assertError(t, err, want)
	})
	t.Run("vertex does not exist", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{Id: 1}
		vertex2 := Vertex{Id: 2}
		g.AddVertex(vertex2)

		err := g.AddEdge(vertex.Id, vertex2.Id, 1)
		want := ErrVertexNotFound

		assertError(t, err, want)
	})
}

func TestRemoveEdge(t *testing.T) {
	t.Run("existing edge", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{Id: 1}
		vertex2 := Vertex{Id: 2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)

		g.AddEdge(vertex.Id, vertex2.Id, 1)
		g.RemoveEdge(vertex.Id, vertex2.Id)
		got := len(g.Edges[vertex.Id])
		want := 0

		assertInt(t, got, want)
	})
	t.Run("not existing edge", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{Id: 1}
		vertex2 := Vertex{Id: 2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)

		err := g.RemoveEdge(vertex.Id, vertex2.Id)
		want := ErrEdgeNotFound

		assertError(t, err, want)
	})
}

func TestNeighbors(t *testing.T) {
	t.Run("existing neighbor", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{Id: 1}
		vertex2 := Vertex{Id: 2}
		g.AddVertex(vertex)
		g.AddVertex(vertex2)
		g.AddEdge(vertex.Id, vertex2.Id, 1)

		got, _ := g.Neighbors(vertex.Id)
		want := []Vertex{vertex2}

		assertSlice(t, got, want)
	})
	t.Run("no neighbor", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{Id: 1}
		g.AddVertex(vertex)

		got, _ := g.Neighbors(vertex.Id)
		want := make([]Vertex, 0)

		assertSlice(t, got, want)
	})
	t.Run("vertex does not exist", func(t *testing.T) {
		g := NewGraph()
		vertex := Vertex{Id: 1}

		_, err := g.Neighbors(vertex.Id)
		want := ErrVertexNotFound

		assertError(t, err, want)
	})
}

func TestAdjacent(t *testing.T) {
	g := NewGraph()
	v1 := Vertex{Id: 1}
	v2 := Vertex{Id: 2}
	g.AddVertex(v1)
	g.AddVertex(v2)
	g.AddEdge(v1.Id, v2.Id, 1)

	t.Run("is adjacent", func(t *testing.T) {
		ok, err := g.Adjacent(v1.Id, v2.Id)
		assertError(t, err, nil)
		if !ok {
			t.Errorf("expected vertices to be adjacent")
		}
	})

	t.Run("not adjacent", func(t *testing.T) {
		v3 := Vertex{Id: 3}
		g.AddVertex(v3)
		ok, err := g.Adjacent(v2.Id, v3.Id)
		assertError(t, err, nil)
		if ok {
			t.Errorf("expected vertices to not be adjacent")
		}
	})

	t.Run("unknown vertex", func(t *testing.T) {
		_, err := g.Adjacent(99, v1.Id)
		assertError(t, err, ErrVertexNotFound)
	})
}


func TestDegree(t *testing.T) {
	g := NewGraph()
	v1 := Vertex{Id: 1}
	v2 := Vertex{Id: 2}
	v3 := Vertex{Id: 3}
	v4 := Vertex{Id: 4}
	t.Run("non existing vertex", func(t *testing.T) {
		_, err := g.Degree(v1.Id)
		assertError(t, err, ErrVertexNotFound)
	})
	t.Run("no neighbors", func(t *testing.T) {
		g.AddVertex(v1)
		g.AddVertex(v2)
		g.AddVertex(v3)
		g.AddVertex(v4)

		got, err := g.Degree(v1.Id)
		assertError(t, err, nil)
		assertInt(t, got, 0)
	})
	t.Run("3 Neighbors", func(t *testing.T) {
		g.AddEdge(v1.Id,v2.Id, 1)
		g.AddEdge(v1.Id,v3.Id, 1)
		g.AddEdge(v1.Id,v4.Id, 1)
		got, err := g.Degree(v1.Id)
		assertError(t, err, nil)
		assertInt(t, got, 3)
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
