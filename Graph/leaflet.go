package graph

// Add these structs to your graph package
import "encoding/json"

type JSONNode struct {
	ID  VertexId `json:"id"`
	Lat float64  `json:"lat"`
	Lon float64  `json:"lon"`
}

type JSONEdge struct {
	From       VertexId `json:"from"`
	To         VertexId `json:"to"`
	IsShortcut bool     `json:"isShortcut"`
}

type JSONGraph struct {
	Nodes []JSONNode `json:"nodes"`
	Edges []JSONEdge `json:"edges"`
}

// ToJSON serializes the graph into a JSON byte slice.
// Make sure this method is part of your Graph type.
func (g *Graph) ToJSON() ([]byte, error) {
	nodeMap := make(map[VertexId]JSONNode)
	jsonGraph := JSONGraph{
		Nodes: make([]JSONNode, 0, len(g.Vertices)),
		Edges: make([]JSONEdge, 0),
	}

	// Convert nodes
	for id, vertex := range g.Vertices {
		node := JSONNode{ID: id, Lat: vertex.Lat, Lon: vertex.Lon}
		jsonGraph.Nodes = append(jsonGraph.Nodes, node)
		nodeMap[id] = node
	}

	// Convert edges
	for sourceId, targets := range g.Edges {
		// Ensure the source node exists in the graph's vertex list
		if _, ok := g.Vertices[sourceId]; !ok {
			continue
		}
		for targetId, edge := range targets {
			// Ensure the target node also exists
			if _, ok := g.Vertices[targetId]; !ok {
				continue
			}
			jsonGraph.Edges = append(jsonGraph.Edges, JSONEdge{
				From:       sourceId,
				To:         targetId,
				IsShortcut: edge.IsShortcut,
			})
		}
	}

	return json.MarshalIndent(jsonGraph, "", "  ")
}
