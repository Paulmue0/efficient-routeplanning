package collection

import (
	"errors"
	"fmt"
)

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

type VertexId int

func (v VertexId) String() string {
	return fmt.Sprintf("ID: %d", v)
}

type Vertex struct {
	Id  VertexId
	Lat float64
	Lon float64
}

func (v Vertex) String() string {
	return fmt.Sprintf("%d", v.Id)
}

type Edge struct {
	Target     VertexId
	Weight     int
	IsShortcut bool
	Via        VertexId
}

func (e Edge) String() string {
	if e.IsShortcut {
		return fmt.Sprintf("Target: %d, Weight: %d, IsShortcutVia %d", e.Target, e.Weight, e.Via)
	}
	return fmt.Sprintf("Target: %d, Weight: %d", e.Target, e.Weight)
}

type Graph struct {
	Vertices map[VertexId]Vertex
	Edges    map[VertexId]map[VertexId]Edge // source -> target -> edge
}

func NewGraph() *Graph {
	return &Graph{
		Vertices: make(map[VertexId]Vertex),
		Edges:    make(map[VertexId]map[VertexId]Edge),
	}
}

func (g *Graph) String() string {
	return fmt.Sprintf("Nodes: %q\n\nEdges: %q", g.Vertices, g.Edges)
}

func (g *Graph) AddVertex(x Vertex) error {
	if _, exists := g.Vertices[x.Id]; exists {
		return ErrVertexAlreadyExists
	}
	g.Vertices[x.Id] = x
	return nil
}

func (g *Graph) RemoveVertex(id VertexId) error {
	if _, exists := g.Vertices[id]; !exists {
		return ErrVertexNotFound
	}

	// Check if vertex has any incoming or outgoing edges
	for src, targets := range g.Edges {
		// Outgoing edge from id
		if src == id {
			return ErrVertexHasEdges
		}

		// Incoming edge to id
		if edge, ok := targets[id]; ok {
			if edge.Target == id {
				return ErrVertexHasEdges
			}
		}
	}

	delete(g.Vertices, id)
	delete(g.Edges, id)

	for src := range g.Edges {
		delete(g.Edges[src], id)
	}

	return nil
}

func (g *Graph) AddEdge(x, y VertexId, weight int, shortcut bool, via VertexId) error {
	if _, ok := g.Vertices[x]; !ok {
		return ErrVertexNotFound
	}
	if _, ok := g.Vertices[y]; !ok {
		return ErrVertexNotFound
	}

	if g.Edges[x] == nil {
		g.Edges[x] = make(map[VertexId]Edge)
	}
	if _, exists := g.Edges[x][y]; exists {
		return ErrEdgeAlreadyExists
	}

	g.Edges[x][y] = Edge{Target: y, Weight: weight, IsShortcut: shortcut, Via: via}
	return nil
}

func (g *Graph) UpdateEdge(x, y VertexId, weight int, shortcut bool, via VertexId) error {
	if g.Edges == nil || g.Edges[x] == nil {
		return ErrEdgeNotFound
	}

	_, exists := g.Edges[x][y]
	// existingEdge, exists := g.Edges[x][y]
	if !exists {
		return ErrEdgeNotFound // Cannot update an edge that does not exist.
	}

	// if weight < existingEdge.Weight {
	g.Edges[x][y] = Edge{Target: y, Weight: weight, IsShortcut: shortcut, Via: via}
	// }

	return nil
}

func (g *Graph) RemoveEdge(x, y VertexId) error {
	if _, exists := g.Edges[x]; !exists {
		return ErrEdgeNotFound
	}
	if _, exists := g.Edges[x][y]; !exists {
		return ErrEdgeNotFound
	}

	delete(g.Edges[x], y)

	if len(g.Edges[x]) == 0 {
		delete(g.Edges, x)
	}
	return nil
}

func (g *Graph) Adjacent(x, y VertexId) (bool, error) {
	if _, ok := g.Vertices[x]; !ok {
		return false, ErrVertexNotFound
	}
	if _, ok := g.Vertices[y]; !ok {
		return false, ErrVertexNotFound
	}

	_, exists := g.Edges[x][y]
	return exists, nil
}

func (g *Graph) Neighbors(x VertexId) ([]Vertex, error) {
	if _, ok := g.Vertices[x]; !ok {
		return nil, ErrVertexNotFound
	}

	targets := g.Edges[x]
	if len(targets) == 0 {
		return nil, nil
	}

	neighbors := make([]Vertex, 0, len(targets))
	for targetId := range targets {
		neighbors = append(neighbors, g.Vertices[targetId])
	}

	return neighbors, nil
}

func (g *Graph) Degree(x VertexId) (int, error) {
	if _, ok := g.Vertices[x]; !ok {
		return 0, ErrVertexNotFound
	}
	return len(g.Edges[x]), nil
}

func (g *Graph) Subgraph(v VertexId) (*Graph, error) {
	if _, exists := g.Vertices[v]; !exists {
		return nil, ErrVertexNotFound
	}
	sub := NewGraph()
	sub.AddVertex(g.Vertices[v])
	neighbors, _ := g.Neighbors(v)

	for _, neighbor := range neighbors {
		sub.AddVertex(neighbor)

		if edge, ok := g.Edges[v][neighbor.Id]; ok {
			sub.AddEdge(v, neighbor.Id, edge.Weight, edge.IsShortcut, edge.Via)
		}
		if edge, ok := g.Edges[neighbor.Id][v]; ok {
			sub.AddEdge(neighbor.Id, v, edge.Weight, edge.IsShortcut, edge.Via)
		}
	}

	for _, from := range neighbors {
		for _, to := range neighbors {
			if from.Id == to.Id {
				continue
			}
			if edge, ok := g.Edges[from.Id][to.Id]; ok {
				sub.AddEdge(from.Id, to.Id, edge.Weight, edge.IsShortcut, edge.Via)
			}
		}
	}

	return sub, nil
}

func (g *Graph) Vertex(id VertexId) (Vertex, error) {
	v, exists := g.Vertices[id]
	if !exists {
		return Vertex{}, ErrVertexNotFound
	}
	return v, nil
}

func (g *Graph) UpdateVertex(id VertexId, v Vertex) error {
	if _, exists := g.Vertices[id]; !exists {
		return ErrVertexNotFound
	}
	g.Vertices[id] = v
	return nil
}

func (g *Graph) NumEdges() int {
	count := 0
	for _, targets := range g.Edges {
		count += len(targets)
	}
	return count
}
