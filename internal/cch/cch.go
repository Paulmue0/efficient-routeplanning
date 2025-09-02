package cch

import graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"

/*
* Customizable Contraction hierarchies is done in three phases
*
* 1. metric independent preprocessing
* 2. metric dependent preprocessing (customization step)
* 3. Query
*
*
*	The idea behind this approach is that any update on the Graph are likely not topological but only affect the
*	metrics of the graph. e.g. weights between vertices.
*
 */

// -------------------- CCH struct ---------------------------------
// upwards graph
// downwards graph
//	ContractionOrder  []graph.VertexId

type CCH struct {
	UpwardsGraph     *graph.Graph
	DownwardsGraph   *graph.Graph
	ContractionOrder []graph.VertexId
	ContractionMap   map[graph.VertexId]int
}

func NewCCH() *CCH {
	ug := graph.NewGraph()
	dg := graph.NewGraph()
	co := make([]graph.VertexId, 0)
	cm := make(map[graph.VertexId]int)

	return &CCH{ug, dg, co, cm}
}

// -------------------- Metric Independent Preprocessing ---------------------------------

// STEP 1.1.: rank order for balanced seperators
//
// Stop when N == 1
// Perform a 2-way split on the graph into B balanced seperators
// Assign the rank order N-1..., N-B to the highest order nodes
//
// Run this on the two resulting graphs again provide parameter for the number of split things on this layer:
//
// 2WaySplit(g0, layer_id: 1)
// 2WaySplit(g1, layer_id: 2)
// 2WaySplit(g2, layer_id: 3)
// 2WaySplit(g3, layer_id: 4)
//
// the 2waysplit funtion assigns the rank order n * layer_id
// this way this step can be done in parallel for each layer

// STEP 1.2: contract the nodes by rank order and insert shortcuts when there is a simple path deleted

// -------------------- Customization / Metric Dependent Preprocessing ---------------------------------
//
// STEP 2.1: basic customization
//
// process all lower triangles in bottom up fashion.
// for each lower triangle
//
//
// 			w
// 		 / \
// 		u   \
//     \- v
// ensure the following:
// weight'(v,w) = min{weight(v,w),weight(v,u) + weight(u,w)}and
// weight'(w,v) = min{weight(w,v),weight(w,u) + weight(u,v)}.
//
//
//
//
// -------------------- Query phase ---------------------------------
//
// Use Elimination Tree for optimization
