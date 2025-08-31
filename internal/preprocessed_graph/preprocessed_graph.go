package preprocessed_graph

import (
	"fmt"
	"os"

	"github.com/PaulMue0/efficient-routeplanning/internal/cch"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
	"github.com/segmentio/parquet-go"
)

// ParquetVertex is a flattened representation of a graph vertex for Parquet.
type ParquetVertex struct {
	ID  int64   `parquet:"id"`
	Lat float64 `parquet:"lat"`
	Lon float64 `parquet:"lon"`
}

// ParquetEdge is a flattened representation of a graph edge for Parquet.
type ParquetEdge struct {
	Source int64 `parquet:"source"`
	Target int64 `parquet:"target"`
	Weight int64 `parquet:"weight"`
	Via    int64 `parquet:"via"`
}

// ParquetContractionMapEntry is a key-value pair for the contraction map.
type ParquetContractionMapEntry struct {
	Key   int64 `parquet:"key"`
	Value int32 `parquet:"value"`
}

// PreprocessedCCHFile holds all the data for a preprocessed CCH graph in a format
// that can be written to a Parquet file.
type PreprocessedCCHFile struct {
	Vertices         []ParquetVertex              `parquet:"vertices"`
	UpwardEdges      []ParquetEdge                `parquet:"upward_edges"`
	DownwardEdges    []ParquetEdge                `parquet:"downward_edges"`
	ContractionOrder []int64                      `parquet:"contraction_order"`
	ContractionMap   []ParquetContractionMapEntry `parquet:"contraction_map"`
}

// FromCCH converts a cch.CCH object into a serializable PreprocessedCCHFile struct.
func FromCCH(cch *cch.CCH) *PreprocessedCCHFile {
	p := &PreprocessedCCHFile{}

	// Vertices
	for _, v := range cch.UpwardsGraph.Vertices {
		p.Vertices = append(p.Vertices, ParquetVertex{ID: int64(v.Id), Lat: v.Lat, Lon: v.Lon})
	}

	// Upward Edges
	for u, edges := range cch.UpwardsGraph.Edges {
		for v, edge := range edges {
			p.UpwardEdges = append(p.UpwardEdges, ParquetEdge{
				Source: int64(u),
				Target: int64(v),
				Weight: int64(edge.Weight),
				Via:    int64(edge.Via),
			})
		}
	}

	// Downward Edges
	for u, edges := range cch.DownwardsGraph.Edges {
		for v, edge := range edges {
			p.DownwardEdges = append(p.DownwardEdges, ParquetEdge{
				Source: int64(u),
				Target: int64(v),
				Weight: int64(edge.Weight),
				Via:    int64(edge.Via),
			})
		}
	}

	// Contraction Order
	for _, id := range cch.ContractionOrder {
		p.ContractionOrder = append(p.ContractionOrder, int64(id))
	}

	// Contraction Map
	for k, v := range cch.ContractionMap {
		p.ContractionMap = append(p.ContractionMap, ParquetContractionMapEntry{Key: int64(k), Value: int32(v)})
	}

	return p
}

// ToCCH converts a PreprocessedCCHFile struct back into a cch.CCH object.
func (p *PreprocessedCCHFile) ToCCH() *cch.CCH {
	cch := cch.NewCCH()

	// Vertices
	for _, pv := range p.Vertices {
		cch.UpwardsGraph.AddVertex(graph.Vertex{Id: graph.VertexId(pv.ID), Lat: pv.Lat, Lon: pv.Lon})
		cch.DownwardsGraph.AddVertex(graph.Vertex{Id: graph.VertexId(pv.ID), Lat: pv.Lat, Lon: pv.Lon})
	}

	// Upward Edges
	for _, pe := range p.UpwardEdges {
		cch.UpwardsGraph.AddEdge(graph.VertexId(pe.Source), graph.VertexId(pe.Target), int(pe.Weight), pe.Via != -1, graph.VertexId(pe.Via))
	}

	// Downward Edges
	for _, pe := range p.DownwardEdges {
		cch.DownwardsGraph.AddEdge(graph.VertexId(pe.Source), graph.VertexId(pe.Target), int(pe.Weight), pe.Via != -1, graph.VertexId(pe.Via))
	}

	// Contraction Order
	for _, id := range p.ContractionOrder {
		cch.ContractionOrder = append(cch.ContractionOrder, graph.VertexId(id))
	}

	// Contraction Map
	for _, entry := range p.ContractionMap {
		cch.ContractionMap[graph.VertexId(entry.Key)] = int(entry.Value)
	}

	return cch
}

// Write saves the PreprocessedCCHFile to a Parquet file.
func (p *PreprocessedCCHFile) Write(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := parquet.NewWriter(file)

	if err := writer.Write(p); err != nil {
		return fmt.Errorf("failed to write parquet data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}