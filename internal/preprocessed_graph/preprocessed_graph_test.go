package preprocessed_graph

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/PaulMue0/efficient-routeplanning/internal/cch"
	"github.com/PaulMue0/efficient-routeplanning/internal/parser"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
	"github.com/segmentio/parquet-go"
)

func TestWriteAndReadCCH(t *testing.T) {
	// 1. Setup: Preprocess the example graph
	orderingPath := "../../data/KaHIP/example.ordering.txt"

	// The parser needs a relative path from the project root, but the test runs in the module dir.
	// So we construct the path relative to the test execution directory.
	fs := os.DirFS("../..")

	net, err := parser.NewNetworkFromFS(fs, "data/RoadNetworks/example.txt")
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}
	absOrderingPath, err := filepath.Abs(orderingPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path for ordering: %v", err)
	}

	cchOriginal := cch.NewCCH()
	err = cchOriginal.Preprocess(net.Network, absOrderingPath)
	if err != nil {
		t.Fatalf("CCH preprocessing failed: %v", err)
	}

	// 2. Convert to serializable format and write to Parquet
	preprocessedFile := FromCCH(cchOriginal)

	tempDir := t.TempDir()
	parquetPath := filepath.Join(tempDir, "test.parquet")

	err = preprocessedFile.Write(parquetPath)
	if err != nil {
		t.Fatalf("Failed to write Parquet file: %v", err)
	}

	// 3. Read from Parquet
	file, err := os.Open(parquetPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	reader := parquet.NewReader(file)

	readData := PreprocessedCCHFile{}
	if err := reader.Read(&readData); err != nil {
		t.Fatalf("Failed to read parquet data: %v", err)
	}

	// 4. Convert back and compare
	cchReconstructed := readData.ToCCH()

	// Custom comparison logic because of maps and unexported fields
	if !reflect.DeepEqual(cchOriginal.ContractionOrder, cchReconstructed.ContractionOrder) {
		t.Errorf("ContractionOrder mismatch: got %v, want %v", cchReconstructed.ContractionOrder, cchOriginal.ContractionOrder)
	}
	if !reflect.DeepEqual(cchOriginal.ContractionMap, cchReconstructed.ContractionMap) {
		t.Errorf("ContractionMap mismatch: got %v, want %v", cchReconstructed.ContractionMap, cchOriginal.ContractionMap)
	}
	if !graphsAreEqual(cchOriginal.UpwardsGraph, cchReconstructed.UpwardsGraph) {
		t.Error("UpwardsGraph is not equal")
	}
	if !graphsAreEqual(cchOriginal.DownwardsGraph, cchReconstructed.DownwardsGraph) {
		t.Error("DownwardsGraph is not equal")
	}
}

// graphsAreEqual is a helper to compare two graph.Graph objects.
func graphsAreEqual(g1, g2 *graph.Graph) bool {
	if len(g1.Vertices) != len(g2.Vertices) {
		fmt.Printf("Vertex count mismatch: %d != %d\n", len(g1.Vertices), len(g2.Vertices))
		return false
	}
	for id, v1 := range g1.Vertices {
		v2, ok := g2.Vertices[id]
		if !ok || !reflect.DeepEqual(v1, v2) {
			fmt.Printf("Vertex mismatch for ID %d: %v != %v\n", id, v1, v2)
			return false
		}
	}

	if len(g1.Edges) != len(g2.Edges) {
		fmt.Printf("Edge source count mismatch: %d != %d\n", len(g1.Edges), len(g2.Edges))
		return false
	}
	for u, edges1 := range g1.Edges {
		edges2, ok := g2.Edges[u]
		if !ok {
			fmt.Printf("Missing source edge map for %d in second graph\n", u)
			return false
		}
		if len(edges1) != len(edges2) {
			fmt.Printf("Edge target count mismatch for source %d: %d != %d\n", u, len(edges1), len(edges2))
			return false
		}
		for v, e1 := range edges1 {
			e2, ok := edges2[v]
			if !ok || e1.Target != e2.Target || e1.Weight != e2.Weight || e1.IsShortcut != e2.IsShortcut || e1.Via != e2.Via {
				fmt.Printf("Edge mismatch for %d->%d: %v != %v\n", u, v, e1, e2)
				return false
			}
		}
	}
	return true
}