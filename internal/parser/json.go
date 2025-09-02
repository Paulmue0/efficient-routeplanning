package parser

import (
	"encoding/json"
	"sort"

	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

type jsonGraph struct {
	Nodes []graph.Vertex `json:"nodes"`
	Edges []jsonEdge     `json:"links"`
}

type jsonEdge struct {
	Source     graph.VertexId `json:"source"`
	Target     graph.VertexId `json:"target"`
	Weight     int            `json:"weight"`
	IsShortcut bool           `json:"is_shortcut"`
	Via        graph.VertexId `json:"via,omitempty"`
}

func ToJSON(g *graph.Graph) ([]byte, error) {
	nodes := make([]graph.Vertex, 0, len(g.Vertices))
	for _, v := range g.Vertices {
		nodes = append(nodes, v)
	}
	// Sort nodes by ID for deterministic output
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Id < nodes[j].Id
	})

	edges := make([]jsonEdge, 0)
	for source, targets := range g.Edges {
		for target, edge := range targets {
			edges = append(edges, jsonEdge{
				Source:     source,
				Target:     target,
				Weight:     edge.Weight,
				IsShortcut: edge.IsShortcut,
				Via:        edge.Via,
			})
		}
	}
	// Sort edges for deterministic output
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Source != edges[j].Source {
			return edges[i].Source < edges[j].Source
		}
		return edges[i].Target < edges[j].Target
	})

	return json.Marshal(&jsonGraph{nodes, edges})
}
