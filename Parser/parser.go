package parser

import (
	"bufio"
	"io"
	"io/fs"
	"strconv"

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
	scanner := bufio.NewScanner(networkFile)

	scanner.Scan()
	numNodes, err := strconv.Atoi(scanner.Text())

	scanner.Scan()
	numEdges, err := strconv.Atoi(scanner.Text())
	check(err)
	network := graph.RoadNetwork{NumNodes: numNodes, NumEdges: numEdges}
	return network, nil
}
