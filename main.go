package main

import (
	"fmt"
	"os"

	parser "github.com/PaulMue0/efficient-routeplanning/Parser"
	"github.com/PaulMue0/efficient-routeplanning/pathfinding"
)

func main() {
	name := "osm1.txt"
	dataDir := "./data/RoadNetworks"
	fileSystem := os.DirFS(dataDir)
	network, err := parser.NewNetworkFromFS(fileSystem, name)
	// originalGraphDOT := network.Network.ToDOT()
	// Export Original Graph
	originalJSON, _ := network.Network.ToJSON()
	os.WriteFile("original.json", originalJSON, 0644)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(network.NumNodes, network.NumEdges, len(network.Network.Vertices), len(network.Network.Edges))

	ch := pathfinding.NewContractionHierarchies()
	ch.Preprocess(network.Network)

	numShortcuts := 0
	for v := range ch.DownwardsGraph.Vertices {
		for _, edge := range ch.DownwardsGraph.Edges[v] {
			if edge.IsShortcut {
				numShortcuts++
			}
		}
	}

	fmt.Println("finished preprocessing. added ", numShortcuts, " shortcuts")
	// UpwardsGraphDOT := ch.UpwardsGraph.ToDOT()
	// os.WriteFile("graph.dot", []byte(originalGraphDOT), 0644)
	// os.WriteFile("UpwardsGraph.dot", []byte(UpwardsGraphDOT), 0644)
	// DownwardsGraphDOT := ch.DownwardsGraph.ToDOT()
	// os.WriteFile("DownwardsGraph.dot", []byte(DownwardsGraphDOT), 0644)
	//
	fmt.Println("Exporting graphs to JSON...")

	// Export Upwards Graph
	upwardsJSON, _ := ch.UpwardsGraph.ToJSON()
	os.WriteFile("upwards.json", upwardsJSON, 0644)

	// Export Downwards Graph
	downwardsJSON, _ := ch.DownwardsGraph.ToJSON()
	os.WriteFile("downwards.json", downwardsJSON, 0644)

	fmt.Println("✅ Done! JSON files are ready.")
}
