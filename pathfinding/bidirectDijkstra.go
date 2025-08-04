package pathfinding

import (
	"container/heap"
	"fmt"
	"math"

	collection "github.com/PaulMue0/efficient-routeplanning/Collection"
	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

// searchContext holds all the state for a single direction of a bidirectional search.
type searchContext struct {
	g     *graph.Graph
	pq    *collection.PriorityQueue[graph.VertexId]
	dists map[graph.VertexId]float64
	preds map[graph.VertexId]graph.VertexId
}

func newSearchContext(g *graph.Graph, startNode graph.VertexId) *searchContext {
	dists := initializeWeights(g, startNode)
	pq := initializePriorityQueue(dists)
	return &searchContext{
		g:     g,
		pq:    pq,
		dists: dists,
		preds: make(map[graph.VertexId]graph.VertexId),
	}
}

func (sc *searchContext) processNextNode(
	oppositeDists map[graph.VertexId]float64,
	shortestPathLength *float64,
	meetNode *graph.VertexId,
) {
	item := heap.Pop(sc.pq).(*collection.Item[graph.VertexId])
	vertex := sc.pq.GetValue(item)
	cost := sc.pq.GetPriority(item)

	if oppositeDist, found := oppositeDists[vertex]; found && !math.IsInf(oppositeDist, 1) {
		if potentialPathLength := cost + oppositeDist; potentialPathLength < *shortestPathLength {
			*shortestPathLength = potentialPathLength
			*meetNode = vertex
		}
	}

	for adjacent, edge := range sc.g.Edges[vertex] {
		newWeight := cost + float64(edge.Weight)
		if newWeight < sc.dists[adjacent] {
			sc.dists[adjacent] = newWeight
			sc.preds[adjacent] = vertex
			sc.pq.UpdatePriority(adjacent, newWeight)
		}
	}
}

func BiDirectionalDijkstraShortestPath(upGraph *graph.Graph, downGraph *graph.Graph, source, target graph.VertexId) ([]graph.VertexId, float64, error) {
	if source == target {
		if _, ok := upGraph.Edges[source]; ok {
			return []graph.VertexId{source}, 0, nil
		}
		return nil, 0, ErrTargetNotReachable
	}

	fwdSearch := newSearchContext(upGraph, source)
	bwdSearch := newSearchContext(downGraph, target)

	currentShortestPath := math.Inf(1)
	var meetNode graph.VertexId

	for fwdSearch.pq.Len() > 0 && bwdSearch.pq.Len() > 0 {
		fwdMinDist := fwdSearch.pq.GetPriority(fwdSearch.pq.Peek())
		bwdMinDist := bwdSearch.pq.GetPriority(bwdSearch.pq.Peek())
		if fwdMinDist+bwdMinDist >= currentShortestPath {
			break
		}

		if fwdMinDist <= bwdMinDist {
			fwdSearch.processNextNode(bwdSearch.dists, &currentShortestPath, &meetNode)
		} else {
			bwdSearch.processNextNode(fwdSearch.dists, &currentShortestPath, &meetNode)
		}
	}
	fmt.Println(currentShortestPath)

	if math.IsInf(currentShortestPath, 1) {
		return nil, 0, ErrTargetNotReachable
	}

	pathFwd, errFwd := buildPath(fwdSearch.preds, source, meetNode)
	if errFwd != nil {
		return nil, 0, errFwd
	}

	pathBwdReversed, errBwd := buildPath(bwdSearch.preds, target, meetNode)
	if errBwd != nil {
		return nil, 0, errBwd
	}

	pathBwd := make([]graph.VertexId, len(pathBwdReversed))
	for i, j := 0, len(pathBwdReversed)-1; j >= 0; i, j = i+1, j-1 {
		pathBwd[i] = pathBwdReversed[j]
	}

	return append(pathFwd, pathBwd[1:]...), currentShortestPath, nil
}
