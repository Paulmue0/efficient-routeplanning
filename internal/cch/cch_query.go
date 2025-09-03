package cch

import (
	"fmt"

	"github.com/PaulMue0/efficient-routeplanning/internal/pathfinding"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

// Query finds the shortest path between source and target using the CCH.
// It performs a bidirectional Dijkstra search on the CCH and then unpacks
// the resulting path to resolve any shortcuts.
func (cch *CCH) Query(source, target graph.VertexId) ([]graph.VertexId, float64, int, error) {
	path, weight, nodesPopped, err := pathfinding.BiDirectionalDijkstraShortestPath(cch.UpwardsGraph, cch.DownwardsGraph, source, target)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("bidirectional Dijkstra failed: %w", err)
	}

	if len(path) == 0 {
		return []graph.VertexId{}, weight, 0, nil
	}

	unpackedPath, err := cch.unpackPath(path)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to unpack path: %w", err)
	}

	return unpackedPath, weight, nodesPopped, nil
}

// unpackPath takes a path containing shortcuts and expands them into the original edges.
func (cch *CCH) unpackPath(path []graph.VertexId) ([]graph.VertexId, error) {
	if len(path) < 2 {
		return path, nil
	}

	fullPath := []graph.VertexId{path[0]}

	for i := 0; i < len(path)-1; i++ {
		u, v := path[i], path[i+1]
		segment, err := cch.unpackEdge(u, v)
		if err != nil {
			return nil, err
		}
		// Append the unpacked segment, skipping the first node to avoid duplication.
		fullPath = append(fullPath, segment[1:]...)
	}

	return fullPath, nil
}

// unpackEdge recursively unpacks a single edge (u, v).
// If the edge is a shortcut, it finds the intermediate node and recursively
// unpacks the two new segments.
func (cch *CCH) unpackEdge(u, v graph.VertexId) ([]graph.VertexId, error) {
	var edge graph.Edge
	var ok bool

	edge, ok = cch.UpwardsGraph.Edges[u][v]
	if !ok {
		edge, ok = cch.DownwardsGraph.Edges[u][v]
		if !ok {
			return nil, fmt.Errorf("no edge found between %d and %d in CCH graphs", u, v)
		}
	}

	if !edge.IsShortcut {
		return []graph.VertexId{u, v}, nil
	}

	via := edge.Via
	path1, err := cch.unpackEdge(u, via)
	if err != nil {
		return nil, err
	}
	path2, err := cch.unpackEdge(via, v)
	if err != nil {
		return nil, err
	}

	// Combine the two unpacked sub-paths.
	// path1 is [u, ..., via], path2 is [via, ..., v].
	// append path2[1:] to path1 to get [u, ..., via, ..., v].
	return append(path1, path2[1:]...), nil
}
