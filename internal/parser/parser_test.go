package parser

import (
	"maps"
	"slices"
	"testing"
	"testing/fstest"

	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
	"github.com/google/go-cmp/cmp"
)

const exampleNetwork = `4
3
0 48.667421 9.244557
1 48.667273 9.244867
2 48.667598 9.244326
3 48.667019 9.245514
0 1
0 2
0 3`

var (
	mockGraph       = graph.NewGraph()
	mockRoadNetwork = graph.RoadNetwork{NumNodes: 4, NumEdges: 3, Network: mockGraph}
)

func TestNewRoadNetwork(t *testing.T) {
	Vertex0 := graph.Vertex{Id: 0, Lat: 48.667421, Lon: 9.244557}
	Vertex1 := graph.Vertex{Id: 1, Lat: 48.667273, Lon: 9.244867}
	Vertex2 := graph.Vertex{Id: 2, Lat: 48.667598, Lon: 9.244326}
	Vertex3 := graph.Vertex{Id: 3, Lat: 48.667019, Lon: 9.245514}

	mockGraph.AddVertex(Vertex0)
	mockGraph.AddVertex(Vertex1)
	mockGraph.AddVertex(Vertex2)
	mockGraph.AddVertex(Vertex3)

	mockGraph.AddEdge(Vertex0.Id, Vertex1.Id, 1, false, -1)
	mockGraph.AddEdge(Vertex0.Id, Vertex2.Id, 1, false, -1)
	mockGraph.AddEdge(Vertex0.Id, Vertex3.Id, 1, false, -1)

	// also reverse edges as it is undirected:
	mockGraph.AddEdge(Vertex1.Id, Vertex0.Id, 1, false, -1)
	mockGraph.AddEdge(Vertex2.Id, Vertex0.Id, 1, false, -1)
	mockGraph.AddEdge(Vertex3.Id, Vertex0.Id, 1, false, -1)
	fs := fstest.MapFS{
		"network-1.txt": {Data: []byte(exampleNetwork)},
	}

	name := slices.Collect(maps.Keys(fs))[0]
	got, err := NewNetworkFromFS(fs, name)
	want := mockRoadNetwork

	assertRoadNetwork(t, got, want)
	assertError(t, err, nil)
}

func assertError(t testing.TB, got error, want error) {
	t.Helper()

	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func assertRoadNetwork(t *testing.T, got graph.RoadNetwork, want graph.RoadNetwork) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("RoadNetwork mismatch (-want +got):\n%s", diff)
	}
}
