package preprocessed_graph

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/PaulMue0/efficient-routeplanning/internal/cch"
	"github.com/PaulMue0/efficient-routeplanning/internal/ch"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

// ParquetVertex is a flattened representation of a graph vertex for Parquet.
type ParquetVertex struct {
	ID  int64
	Lat float64
	Lon float64
}

// ParquetEdge is a flattened representation of a graph edge for Parquet.
type ParquetEdge struct {
	Source int64
	Target int64
	Weight int64
	Via    int64
}

// ParquetContractionMapEntry is a key-value pair for the contraction map.
type ParquetContractionMapEntry struct {
	Key   int64
	Value int32
}

// PreprocessedCCHFile holds all the data for a preprocessed CCH graph in a format
// that can be written to a gob file.
type PreprocessedCCHFile struct {
	Vertices         []ParquetVertex
	UpwardEdges      []ParquetEdge
	DownwardEdges    []ParquetEdge
	ContractionOrder []int64
	ContractionMap   []ParquetContractionMapEntry
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

// Write saves the PreprocessedCCHFile to a gob file.
func (p *PreprocessedCCHFile) Write(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(p); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	return nil
}

// ReadCCH reads a PreprocessedCCHFile from a gob file.
func ReadCCH(path string) (*PreprocessedCCHFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var p PreprocessedCCHFile
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&p); err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	return &p, nil
}

// GraphData is a serializable representation of graph.Graph
type GraphData struct {
	Vertices []ParquetVertex
	Edges    []ParquetEdge
}

// PreprocessedCHFile holds all the data for a preprocessed CH graph in a format
// that can be written to a gob file.
type PreprocessedCHFile struct {
	ContractionOrder []int64
	UpwardsGraph     *GraphData
	DownwardsGraph   *GraphData
}

// FromCH converts a ch.ContractionHierarchies object into a serializable PreprocessedCHFile struct.
func FromCH(ch *ch.ContractionHierarchies) *PreprocessedCHFile {
	p := &PreprocessedCHFile{}

	// Contraction Order
	for _, id := range ch.ContractionOrder {
		p.ContractionOrder = append(p.ContractionOrder, int64(id))
	}

	// Upwards Graph
	ug := &GraphData{}
	for _, v := range ch.UpwardsGraph.Vertices {
		ug.Vertices = append(ug.Vertices, ParquetVertex{ID: int64(v.Id), Lat: v.Lat, Lon: v.Lon})
	}
	for u, edges := range ch.UpwardsGraph.Edges {
		for v, edge := range edges {
			ug.Edges = append(ug.Edges, ParquetEdge{
				Source: int64(u),
				Target: int64(v),
				Weight: int64(edge.Weight),
				Via:    int64(edge.Via),
			})
		}
	}
	p.UpwardsGraph = ug

	// Downwards Graph
	dg := &GraphData{}
	for _, v := range ch.DownwardsGraph.Vertices {
		dg.Vertices = append(dg.Vertices, ParquetVertex{ID: int64(v.Id), Lat: v.Lat, Lon: v.Lon})
	}
	for u, edges := range ch.DownwardsGraph.Edges {
		for v, edge := range edges {
			dg.Edges = append(dg.Edges, ParquetEdge{
				Source: int64(u),
				Target: int64(v),
				Weight: int64(edge.Weight),
				Via:    int64(edge.Via),
			})
		}
	}
	p.DownwardsGraph = dg

	return p
}

// ToCH converts a PreprocessedCHFile struct back into a ch.ContractionHierarchies object.
func (p *PreprocessedCHFile) ToCH() *ch.ContractionHierarchies {
	ch := ch.NewContractionHierarchies()

	// Contraction Order
	for _, id := range p.ContractionOrder {
		ch.ContractionOrder = append(ch.ContractionOrder, graph.VertexId(id))
	}

	// Upwards Graph
	for _, pv := range p.UpwardsGraph.Vertices {
		ch.UpwardsGraph.AddVertex(graph.Vertex{Id: graph.VertexId(pv.ID), Lat: pv.Lat, Lon: pv.Lon})
	}
	for _, pe := range p.UpwardsGraph.Edges {
		ch.UpwardsGraph.AddEdge(graph.VertexId(pe.Source), graph.VertexId(pe.Target), int(pe.Weight), pe.Via != -1, graph.VertexId(pe.Via))
	}

	// Downwards Graph
	for _, pv := range p.DownwardsGraph.Vertices {
		ch.DownwardsGraph.AddVertex(graph.Vertex{Id: graph.VertexId(pv.ID), Lat: pv.Lat, Lon: pv.Lon})
	}
	for _, pe := range p.DownwardsGraph.Edges {
		ch.DownwardsGraph.AddEdge(graph.VertexId(pe.Source), graph.VertexId(pe.Target), int(pe.Weight), pe.Via != -1, graph.VertexId(pe.Via))
	}

	return ch
}

// WriteCH saves the PreprocessedCHFile to a gob file.
func (p *PreprocessedCHFile) WriteCH(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(p); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	return nil
}

// ReadCHFile reads a PreprocessedCHFile from a gob file.
func ReadCHFile(path string) (*PreprocessedCHFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var p PreprocessedCHFile
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&p); err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	return &p, nil
}