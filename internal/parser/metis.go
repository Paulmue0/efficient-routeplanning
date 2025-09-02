package parser

import (
	"bufio"
	"fmt"
	"io"
	"sort"

	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

// ToMetis converts a graph to the METIS text format and writes it to the provided writer.
// The format consists of a single header line followed by a line for each vertex.
// The header contains the number of vertices and the number of undirected edges.
// Each subsequent line lists the 1-based indices of the neighbors for a vertex.
func ToMetis(g *graph.Graph, w io.Writer) error {
	bufferedWriter := bufio.NewWriter(w)
	defer bufferedWriter.Flush()

	// 1. Create a deterministic mapping from graph VertexId to METIS's 1-based index.
	nodeIDs := make([]graph.VertexId, 0, len(g.Vertices))
	for id := range g.Vertices {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Slice(nodeIDs, func(i, j int) bool {
		return nodeIDs[i] < nodeIDs[j]
	})

	graphIdToMetisId := make(map[graph.VertexId]int, len(nodeIDs))
	for i, id := range nodeIDs {
		// Use a 1-based index for the METIS format.
		graphIdToMetisId[id] = i + 1
	}

	numNodes := len(g.Vertices)
	adj := make([][]int, numNodes)
	numEdges := 0

	// 2. Build an undirected adjacency list and count the number of undirected edges.
	tempAdj := make([]map[int]struct{}, numNodes)
	for i := range tempAdj {
		tempAdj[i] = make(map[int]struct{})
	}

	for srcID, targets := range g.Edges {
		for tgtID := range targets {
			srcMetisID := graphIdToMetisId[srcID]
			tgtMetisID := graphIdToMetisId[tgtID]

			// Skip self-loops as they are not supported by the format.
			if srcMetisID == tgtMetisID {
				continue
			}

			// Add edges in both directions to represent an undirected graph.
			tempAdj[srcMetisID-1][tgtMetisID] = struct{}{}
			tempAdj[tgtMetisID-1][srcMetisID] = struct{}{}
		}
	}

	// Convert the map-based adjacency list to a slice-based one for sorting and writing.
	totalDegree := 0
	for i, neighborsMap := range tempAdj {
		neighbors := make([]int, 0, len(neighborsMap))
		for neighbor := range neighborsMap {
			neighbors = append(neighbors, neighbor)
		}
		// Sort neighbors for a deterministic file format.
		sort.Slice(neighbors, func(a, b int) bool {
			return neighbors[a] < neighbors[b]
		})
		adj[i] = neighbors
		totalDegree += len(neighbors)
	}

	// In an undirected graph, the number of edges is the sum of all degrees divided by 2.
	numEdges = totalDegree / 2

	// 3. Write the header information.
	// The METIS text format header is: <number of vertices> <number of edges>
	if _, err := fmt.Fprintf(bufferedWriter, "%d %d\n", numNodes, numEdges); err != nil {
		return err
	}

	// 4. Write the adjacency list.
	// Each line corresponds to a vertex and lists its neighbors.
	for _, neighbors := range adj {
		for i, neighbor := range neighbors {
			if i > 0 {
				if _, err := fmt.Fprint(bufferedWriter, " "); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprint(bufferedWriter, neighbor); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(bufferedWriter, "\n"); err != nil {
			return err
		}
	}

	return nil
}
