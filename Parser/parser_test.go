package parser_test

import (
	"maps"
	"reflect"
	"slices"
	"testing"
	"testing/fstest"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
	parser "github.com/PaulMue0/efficient-routeplanning/Parser"
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
	Vertex0 := graph.Vertex{Id: 3, Lat: 48.667019, Lon: 9.245514}
	Vertex1 := graph.Vertex{Id: 2, Lat: 48.667598, Lon: 9.244326}
	Vertex2 := graph.Vertex{Id: 1, Lat: 48.667273, Lon: 9.244867}
	Vertex3 := graph.Vertex{Id: 0, Lat: 48.667421, Lon: 9.244557}

	Edge0 := graph.Edge{Target: Vertex1, Weight: 1}
	Edge1 := graph.Edge{Target: Vertex2, Weight: 1}
	Edge2 := graph.Edge{Target: Vertex3, Weight: 1}

	mockGraph.AddVertex(Vertex0)
	mockGraph.AddVertex(Vertex1)
	mockGraph.AddVertex(Vertex2)
	mockGraph.AddVertex(Vertex3)

	mockGraph.AddEdge(Vertex0, Edge0)
	mockGraph.AddEdge(Vertex0, Edge1)
	mockGraph.AddEdge(Vertex0, Edge2)
	fs := fstest.MapFS{
		"network-1.txt": {Data: []byte(exampleNetwork)},
	}

	name := slices.Collect(maps.Keys(fs))[0]
	got, err := parser.NewNetworkFromFS(fs, name)
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
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
