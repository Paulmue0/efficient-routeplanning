package main

import (
	"fmt"
	"os"

	parser "github.com/PaulMue0/efficient-routeplanning/Parser"
)

func main() {
	name := "osm1.txt"
	dataDir := "./data/RoadNetworks"
	fileSystem := os.DirFS(dataDir)
	network, err := parser.NewNetworkFromFS(fileSystem, name)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(network.NumNodes, network.NumEdges, len(network.Network.Vertices), len(network.Network.Edges))
}
