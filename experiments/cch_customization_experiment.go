package experiments

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PaulMue0/efficient-routeplanning/internal/parser"
	"github.com/PaulMue0/efficient-routeplanning/internal/preprocessed_graph"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

type CCHCustomizationExperimentResult struct {
	GraphName                 string
	OriginalCustomizationTime time.Duration
	AvgRandomCustomizationTime  time.Duration
}

func RunCCHCustomizationExperiment() {
	preprocessedDir := "./data/preprocessed"
	roadNetworksDir := "./data/RoadNetworks"
	resultsPath := "./cch_customization_experiment_results.csv"
	numRandomRuns := 5

	files, err := os.ReadDir(preprocessedDir)
	if err != nil {
		log.Fatalf("failed to read preprocessed directory: %v", err)
	}

	var results []CCHCustomizationExperimentResult

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "cch_osm") && strings.HasSuffix(file.Name(), ".gob") {
			graphName := strings.TrimSuffix(strings.TrimPrefix(file.Name(), "cch_"), ".gob") + ".txt"
			log.Printf("Processing graph: %s", graphName)

			// Load original graph
			fileSystem := os.DirFS(roadNetworksDir)
			originalNetwork, err := parser.NewNetworkFromFS(fileSystem, graphName)
			if err != nil {
				log.Printf("failed to load original graph %s: %v", graphName, err)
				continue
			}

			// Load CCH graph
			cchFilePath := filepath.Join(preprocessedDir, file.Name())
			preprocessedFile, err := preprocessed_graph.ReadCCH(cchFilePath)
			if err != nil {
				log.Printf("failed to read preprocessed cch for %s: %v", graphName, err)
				continue
			}
			cchInstance := preprocessedFile.ToCCH()

			// --- Original Weights Run ---
			start := time.Now()
			err = cchInstance.Customize(originalNetwork.Network)
			if err != nil {
				log.Printf("failed to customize with original weights for %s: %v", graphName, err)
				continue
			}
			originalTime := time.Since(start)

			// --- Random Weights Runs ---
			var totalRandomTime time.Duration
			for i := 0; i < numRandomRuns; i++ {
				// Create a graph with random weights
				randomWeightGraph := createRandomWeightGraph(originalNetwork.Network)

				// Reload CCH instance to have a fresh start
				cchInstance := preprocessedFile.ToCCH()

				start := time.Now()
				err = cchInstance.Customize(randomWeightGraph)
				if err != nil {
					log.Printf("failed to customize with random weights for %s (run %d): %v", graphName, i+1, err)
					continue
				}
				randomTime := time.Since(start)
				totalRandomTime += randomTime
			}
			avgRandomTime := totalRandomTime / time.Duration(numRandomRuns)

			result := CCHCustomizationExperimentResult{
				GraphName:                 graphName,
				OriginalCustomizationTime: originalTime,
				AvgRandomCustomizationTime:  avgRandomTime,
			}
			results = append(results, result)

			log.Printf("Finished processing %s", graphName)
		}
	}

	// Write results to CSV
	csvFile, err := os.Create(resultsPath)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	headers := []string{"Graph", "OriginalCustomizationTime(ms)", "AvgRandomCustomizationTime(ms)"}
	writer.Write(headers)

	for _, result := range results {
		row := []string{
			result.GraphName,
			fmt.Sprintf("%.3f", float64(result.OriginalCustomizationTime.Nanoseconds())/1e6),
			fmt.Sprintf("%.3f", float64(result.AvgRandomCustomizationTime.Nanoseconds())/1e6),
		}
		writer.Write(row)
	}

	log.Printf("CCH customization experiment results written to %s", resultsPath)
}

func createRandomWeightGraph(original *graph.Graph) *graph.Graph {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomGraph := graph.NewGraph()

	for _, vertex := range original.Vertices {
		randomGraph.AddVertex(vertex)
	}

	for u, edges := range original.Edges {
		for v := range edges {
			if u < v { // Add each edge only once
				randomWeight := r.Intn(1000) + 1 // [1, 1000]
				randomGraph.AddEdge(u, v, randomWeight, false, -1)
				randomGraph.AddEdge(v, u, randomWeight, false, -1)
			}
		}
	}

	return randomGraph
}
