package pathfinding

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

func (c *CCH) Preprocess(g *graph.Graph, orderingFilePath string) error {
	fmt.Printf("Building CCH using ordering file: %s\n", orderingFilePath)

	if err := c.initializeContraction(g, orderingFilePath); err != nil {
		return fmt.Errorf("failed to initialize contraction: %w", err)
	}

	if err := c.initializeGraphsWithVertices(g); err != nil {
		return fmt.Errorf("failed to initialize graphs with vertices: %w", err)
	}

	if err := c.populateGraphsWithEdges(g); err != nil {
		return fmt.Errorf("failed to populate graphs with edges: %w", err)
	}

	if err := c.addShortcuts(); err != nil {
		return fmt.Errorf("failed to add shortcuts: %w", err)
	}

	return nil
}

// initializeGraphsWithVertices adds all vertices from the original graph to
// both the upwards and downwards CCH graphs.
func (c *CCH) initializeGraphsWithVertices(g *graph.Graph) error {
	for _, v := range g.Vertices {
		if err := c.UpwardsGraph.AddVertex(v); err != nil {
			return fmt.Errorf("failed to add vertex to upwards graph: %w", err)
		}
		if err := c.DownwardsGraph.AddVertex(v); err != nil {
			return fmt.Errorf("failed to add vertex to downwards graph: %w", err)
		}
	}
	return nil
}

// populateGraphsWithEdges iterates through the original graph's edges and adds them
// to the upwards and downwards graph based on the contraction order.
func (c *CCH) populateGraphsWithEdges(g *graph.Graph) error {
	seen := make(map[[2]graph.VertexId]bool)
	for uID, outgoingEdges := range g.Edges {
		for vID, edge := range outgoingEdges {
			key := [2]graph.VertexId{uID, vID}
			if uID > vID {
				key = [2]graph.VertexId{vID, uID}
			}
			if seen[key] {
				continue
			}
			seen[key] = true

			// If u's contraction rank is lower than v's, it's an upwards edge.
			if c.ContractionMap[uID] < c.ContractionMap[vID] {
				if err := c.UpwardsGraph.AddEdge(uID, vID, edge.Weight, false, -1); err != nil {
					return fmt.Errorf("failed to add edge (%d -> %d) to upwards graph: %w", uID, vID, err)
				}
				if err := c.DownwardsGraph.AddEdge(vID, uID, edge.Weight, false, -1); err != nil {
					return fmt.Errorf("failed to add edge (%d -> %d) to DownwardsGraph graph: %w", vID, uID, err)
				}
			} else {
				// Otherwise, it's a downwards edge.
				if err := c.DownwardsGraph.AddEdge(uID, vID, edge.Weight, false, -1); err != nil {
					return fmt.Errorf("failed to add edge (%d -> %d) to downwards graph: %w", uID, vID, err)
				}
				if err := c.UpwardsGraph.AddEdge(vID, uID, edge.Weight, false, -1); err != nil {
					return fmt.Errorf("failed to add edge (%d -> %d) to downwards graph: %w", vID, uID, err)
				}
			}
		}
	}
	return nil
}

func (c *CCH) addShortcuts() error {
	for _, id := range c.ContractionOrder {
		higherRankedNeighbors, err := c.UpwardsGraph.Neighbors(id)
		if err != nil {
			return fmt.Errorf("failed to get up-neighbors for vertex %d: %w", id, err)
		}

		// Connect all higher-ranked neighbors to create shortcuts for paths through 'id'.
		for i := 0; i < len(higherRankedNeighbors); i++ {
			for j := i + 1; j < len(higherRankedNeighbors); j++ {
				uNode := higherRankedNeighbors[i]
				wNode := higherRankedNeighbors[j]
				u, w := uNode.Id, wNode.Id

				// Determine the direction of the shortcut based on contraction rank.
				// The edge in the UpwardsGraph always points from lower rank to higher rank.
				var start, end graph.VertexId
				if c.ContractionMap[u] < c.ContractionMap[w] {
					start, end = u, w
				} else {
					start, end = w, u
				}

				exists, err := c.UpwardsGraph.Adjacent(start, end)
				if err != nil {
					return fmt.Errorf("failed to check adjacency between %d and %d: %w", start, end, err)
				}
				if !exists {
					// The weight is set to infinity to be updated later by metric-dependent steps.
					if err := c.UpwardsGraph.AddEdge(start, end, int(math.Inf(1)), true, id); err != nil {
						return fmt.Errorf("failed to add shortcut (%d -> %d) to upwards graph: %w", start, end, err)
					}
					// The DownwardsGraph is the reverse of the UpwardsGraph.
					if err := c.DownwardsGraph.AddEdge(end, start, int(math.Inf(1)), true, id); err != nil {
						return fmt.Errorf("failed to add shortcut (%d -> %d) to downwards graph: %w", end, start, err)
					}
				}
			}
		}
	}
	return nil
}

func (c *CCH) initializeContraction(g *graph.Graph, orderingFilePath string) error {
	// Re-create the mapping from METIS's 1-based ID to the original graph's VertexId.
	nodeIDs := make([]graph.VertexId, 0, len(g.Vertices))
	for id := range g.Vertices {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Slice(nodeIDs, func(i, j int) bool {
		return nodeIDs[i] < nodeIDs[j]
	})

	metisIdToGraphId := make(map[int]graph.VertexId, len(nodeIDs))
	for i, id := range nodeIDs {
		metisIdToGraphId[i+1] = id
	}

	file, err := os.Open(orderingFilePath)
	if err != nil {
		return fmt.Errorf("failed to open ordering file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	metisOrdering := make([]int, 0, len(g.Vertices))
	for scanner.Scan() {
		line := scanner.Text()
		metisID, err := strconv.Atoi(line)
		if err != nil {
			return fmt.Errorf("invalid integer found in ordering file: %w", err)
		}
		metisOrdering = append(metisOrdering, metisID)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading ordering file: %w", err)
	}

	if len(metisOrdering) != len(g.Vertices) {
		return fmt.Errorf("mismatch in node count: graph has %d nodes, but ordering file has %d entries",
			len(g.Vertices), len(metisOrdering))
	}

	originalNodeOrdering := make([]graph.VertexId, len(metisOrdering))
	for i, metisID := range metisOrdering {
		originalID, ok := metisIdToGraphId[metisID]
		if !ok {
			return fmt.Errorf("METIS ID %d from file not found in graph mapping", metisID)
		}
		originalNodeOrdering[i] = originalID
	}

	// The contraction order is the reverse of the node ordering.
	contractionOrder := make([]graph.VertexId, len(originalNodeOrdering))
	contractionMap := make(map[graph.VertexId]int)
	for i, nodeID := range originalNodeOrdering {
		contractionOrder[len(originalNodeOrdering)-1-i] = nodeID
		contractionMap[nodeID] = len(originalNodeOrdering) - 1 - i
	}

	c.ContractionOrder = contractionOrder
	c.ContractionMap = contractionMap
	return nil
}
