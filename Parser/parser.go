package parser

import (
	"bufio"
	"io"
	"io/fs"
	"strconv"
	"strings"

	graph "github.com/PaulMue0/efficient-routeplanning/Graph"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func NewNetworkFromFS(fileSystem fs.FS, name string) (graph.RoadNetwork, error) {
	file, err := fileSystem.Open(name)

	check(err)

	defer file.Close()
	return newNetwork(file)
}

func newNetwork(networkFile io.Reader) (graph.RoadNetwork, error) {
	g := graph.NewGraph()
	scanner := bufio.NewScanner(networkFile)

	scanner.Scan()
	numNodes, err := strconv.Atoi(scanner.Text())
	check(err)

	scanner.Scan()
	numEdges, err := strconv.Atoi(scanner.Text())
	check(err)

	for scanner.Scan() {
		properties := strings.Fields(scanner.Text())
		if len(properties) == 3 {
			id, _ := strconv.Atoi(properties[0])
			lat, _ := strconv.ParseFloat(properties[1], 64)
			lon, _ := strconv.ParseFloat(properties[2], 64)
			vertex := graph.Vertex{Id: graph.VertexId(id), Lat: lat, Lon: lon}
			g.AddVertex(vertex)
		}
		if len(properties) == 2 {
			sourceId, _ := strconv.Atoi(properties[0])
			targetId, _ := strconv.Atoi(properties[1])
			source, _ := g.Vertex(graph.VertexId(sourceId))
			target, _ := g.Vertex(graph.VertexId(targetId))

			g.AddEdge(source.Id, target.Id, 1)
		}
	}

	network := graph.RoadNetwork{NumNodes: numNodes, NumEdges: numEdges, Network: g}
	return network, nil
}
