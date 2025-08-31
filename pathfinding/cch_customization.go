package pathfinding

import (
	"fmt"
	"math"
	"sort"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

func (cch *CCH) Customize(originalGraph *graph.Graph) error {
	if err := cch.Respecting(originalGraph); err != nil {
		return fmt.Errorf("failed to perform repecting %w", err)
	}

	if err := cch.basicCustomization(); err != nil {
		return fmt.Errorf("failed to perform basic customization: %w", err)
	}
	return nil
}

func (cch *CCH) Respecting(originalGraph *graph.Graph) error {
	for v := range cch.UpwardsGraph.Vertices {
		for w, edge := range cch.UpwardsGraph.Edges[v] {
			originalEdge, exists := originalGraph.Edges[v][w]

			var (
				newWeight  int
				isShortcut bool
				viaNode    graph.VertexId
			)

			if exists {
				newWeight = originalEdge.Weight
				isShortcut = false
				viaNode = -1
			} else {
				newWeight = int(math.Inf(1))
				isShortcut = true
				viaNode = edge.Via
			}

			if err := cch.UpwardsGraph.UpdateEdge(v, w, newWeight, isShortcut, viaNode); err != nil {
				return fmt.Errorf("failed to update upwards graph for edge %d->%d: %w", v, w, err)
			}

			if err := cch.DownwardsGraph.UpdateEdge(w, v, newWeight, isShortcut, viaNode); err != nil {
				return fmt.Errorf("failed to update downwards graph for edge %d->%d: %w", w, v, err)
			}
		}
	}
	return nil
}

func (cch *CCH) basicCustomization() error {
	if cch == nil {
		return fmt.Errorf("cch is nil")
	}

	for _, uId := range cch.ContractionOrder {
		upwardsNeighbors, err := cch.UpwardsGraph.Neighbors(uId)
		if err != nil {
			return fmt.Errorf("failed to get neighbors for node %d: %w", uId, err)
		}

		sort.Slice(upwardsNeighbors, func(i, j int) bool {
			rank_w_i, ok := cch.ContractionMap[upwardsNeighbors[i].Id]
			if !ok {
				// This panic is justified if the data structure integrity is guaranteed.
				// Otherwise, a more graceful error handling might be needed.
				panic(fmt.Sprintf("node ID %d not found in ContractionMap", upwardsNeighbors[i].Id))
			}
			rank_w_j, ok := cch.ContractionMap[upwardsNeighbors[j].Id]
			if !ok {
				panic(fmt.Sprintf("node ID %d not found in ContractionMap", upwardsNeighbors[j].Id))
			}
			return rank_w_i < rank_w_j
		})

		for _, v := range upwardsNeighbors {
			for _, w := range upwardsNeighbors {
				rankV, okV := cch.ContractionMap[v.Id]
				rankW, okW := cch.ContractionMap[w.Id]
				if !okV || !okW {
					return fmt.Errorf("node ID lookup failed for v=%d or w=%d", v.Id, w.Id)
				}
				if rankV >= rankW {
					continue
				}

				adjUV, _ := cch.UpwardsGraph.Adjacent(uId, v.Id)
				adjVW, _ := cch.UpwardsGraph.Adjacent(v.Id, w.Id)

				if !adjUV || !adjVW {
					continue
				}

				edgeUV, okUV := cch.UpwardsGraph.Edges[uId][v.Id]
				edgeUW, okUW := cch.UpwardsGraph.Edges[uId][w.Id]
				edgeWU, okWU := cch.DownwardsGraph.Edges[w.Id][uId]
				edgeVU, okVU := cch.DownwardsGraph.Edges[v.Id][uId]

				if !okUV || !okUW || !okWU || !okVU {
					return fmt.Errorf("missing required edges for triangle u=%d, v=%d, w=%d", uId, v.Id, w.Id)
				}

				existingUpEdge, okUp := cch.UpwardsGraph.Edges[v.Id][w.Id]
				if !okUp {
					return fmt.Errorf("missing edge (%d, %d) in upwards graph", v.Id, w.Id)
				}
				existingDownEdge, okDown := cch.DownwardsGraph.Edges[w.Id][v.Id]
				if !okDown {
					return fmt.Errorf("missing edge (%d, %d) in downwards graph", w.Id, v.Id)
				}

				newUpwardsWeight := min(existingUpEdge.Weight, edgeVU.Weight+edgeUW.Weight)
				newDownwardsWeight := min(existingDownEdge.Weight, edgeUV.Weight+edgeWU.Weight)

				if newUpwardsWeight >= existingUpEdge.Weight || newDownwardsWeight >= existingDownEdge.Weight {
					continue
				}

				if err := cch.UpwardsGraph.UpdateEdge(v.Id, w.Id, newUpwardsWeight, true, uId); err != nil {
					return fmt.Errorf("failed to update upwards edge (%d, %d): %w", v.Id, w.Id, err)
				}
				if err := cch.DownwardsGraph.UpdateEdge(w.Id, v.Id, newDownwardsWeight, true, uId); err != nil {
					return fmt.Errorf("failed to update downwards edge (%d, %d): %w", w.Id, v.Id, err)
				}
			}
		}
	}
	return nil
}
