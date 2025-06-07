package graph

import (
	"errors"
	"fmt"
	"slices"
)

/*
*   - Adjacent(G, x, y): tests whether there is an edge from the vertex x to the vertex y;
*   - Neighbors(G, x): lists all vertices y such that there is an edge from the vertex x to the vertex y;
*   - Add_vertex(G, x): adds the vertex x, if it is not there;
*   - Remove_vertex(G, x): removes the vertex x, if it is there;
*   - Add_edge(G, x, y, z): adds the edge z from the vertex x to the vertex y, if it is not there;
*   - Remove_edge(G, x, y): removes the edge from the vertex x to the vertex y, if it is there;
*   - Get_vertex_value(G, x): returns the value associated with the vertex x;
*   - Set_vertex_value(G, x, v): sets the value associated with the vertex x to v.
 */

var (
	ErrVertexNotFound      = errors.New("vertex not found")
	ErrVertexAlreadyExists = errors.New("vertex already exists")
	ErrEdgeNotFound        = errors.New("edge not found")
	ErrEdgeAlreadyExists   = errors.New("edge already exists")
	ErrEdgeCreatesCycle    = errors.New("edge would create a cycle")
	ErrVertexHasEdges      = errors.New("vertex has edges")
)

type RoadNetwork struct {
	NumNodes int
	NumEdges int
	Network  *Graph
}

func NewRoadNetwork() *RoadNetwork {
	return &RoadNetwork{}
}

type Graph struct {
	AdjacencyList map[Vertex][]Edge
}

func NewGraph() *Graph {
	return &Graph{make(map[Vertex][]Edge)}
}

// *   - Adjacent(G, x, y): tests whether there is an edge from the vertex x to the vertex y;
func (g *Graph) Adjacent(x, y Vertex) (bool, error) {
	edges, err := g.Search(x)
	if err != nil {
		return false, err
	}
	for _, edge := range edges {
		if edge.Target == y {
			return true, nil
		}
	}
	return false, nil
}

// *   - Neighbors(G, x): lists all vertices y such that there is an edge from the vertex x to the vertex y;
// return error when vertex does not exist
func (g *Graph) Neighbors(x Vertex) ([]Vertex, error) {
	edges, err := g.Search(x)
	if err != nil {
		return nil, err
	}
	if len(edges) == 0 {
		return nil, nil
	}
	vertices := make([]Vertex, 0)

	for _, edge := range edges {
		vertices = append(vertices, edge.Target)
	}
	return vertices, nil
}

func (g *Graph) Search(v Vertex) ([]Edge, error) {
	edges, ok := g.AdjacencyList[v]
	if !ok {
		return edges, ErrVertexNotFound
	}
	return g.AdjacencyList[v], nil
}

func (g *Graph) ExistsEdge(v Vertex, e Edge) (bool, error) {
	edges, err := g.Search(v)
	if err == ErrVertexNotFound {
		return false, ErrVertexNotFound
	}
	return slices.Contains(edges, e), nil
}

// *   - Add_vertex(G, x): adds the vertex x, if it is not there;
func (g *Graph) AddVertex(x Vertex) error {
	if _, exists := g.AdjacencyList[x]; exists {
		return ErrVertexAlreadyExists
	}
	g.AdjacencyList[x] = []Edge{}
	return nil
}

// *   - Remove_edge(G, x, y): removes the edge from the vertex x to the vertex y, if it is there;
func (g *Graph) RemoveVertex(x Vertex) error {
	if _, ok := g.AdjacencyList[x]; !ok {
		return ErrVertexNotFound
	}
	for v, edges := range g.AdjacencyList {
		newEdges := make([]Edge, 0)
		for _, e := range edges {
			if e.Target != x {
				newEdges = append(newEdges, e)
			}
		}
		g.AdjacencyList[v] = newEdges
	}

	delete(g.AdjacencyList, x)
	return nil
}

func (g *Graph) AddEdge(source Vertex, edge Edge) error {
	exists, err := g.ExistsEdge(source, edge)
	if err != nil {
		return err
	}
	if exists {
		return ErrEdgeAlreadyExists
	}
	g.AdjacencyList[source] = append(g.AdjacencyList[source], edge)
	return nil
}

func (g *Graph) RemoveEdge(source Vertex, edge Edge) error {
	exists, err := g.ExistsEdge(source, edge)
	if err != nil {
		return err
	}
	if !exists {
		return ErrEdgeNotFound
	}
	index := slices.Index(g.AdjacencyList[source], edge)
	g.AdjacencyList[source] = slices.Delete(g.AdjacencyList[source], index, index+1)
	return nil
}

func (g *Graph) Degree(vertex Vertex) (int, error) {
	edges, err := g.Search(vertex)
	if err != nil {
		return 0, err
	}
	return len(edges), nil
}

type Vertex struct {
	Id  VertexId
	Lat float64
	Lon float64
}

func (v Vertex) String() string {
	return fmt.Sprintf("ID: %d", v.Id)
}

type VertexId int

type Edge struct {
	Target Vertex
	Weight int
}

func (e Edge) String() string {
	return fmt.Sprintf("Source: %q, Weight: %d", e.Target, e.Weight)
}
