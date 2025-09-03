package experiments

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PaulMue0/efficient-routeplanning/internal/parser"
	"github.com/PaulMue0/efficient-routeplanning/internal/pathfinding"
	"github.com/PaulMue0/efficient-routeplanning/internal/preprocessed_graph"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

type CCHQueryExperimentResult struct {
	GraphName       string
	AvgDijkstraTime time.Duration
	AvgCCHQueryTime time.Duration
	Mismatches      int
}

func RunCCHQueryExperiment() {
	preprocessedDir := "./data/preprocessed"
	roadNetworksDir := "./data/RoadNetworks"
	resultsPath := "./cch_query_experiment_results.csv"
	numQueries := 100

	files, err := os.ReadDir(preprocessedDir)
	if err != nil {
		log.Fatalf("failed to read preprocessed directory: %v", err)
	}

	var results []CCHQueryExperimentResult

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

			// Customize CCH with original weights
			err = cchInstance.Customize(originalNetwork.Network)
			if err != nil {
				log.Printf("failed to customize cch for %s: %v", graphName, err)
				continue
			}

			// Get vertices for random queries
			var vertices []graph.VertexId
			for _, v := range originalNetwork.Network.Vertices {
				vertices = append(vertices, v.Id)
			}
			if len(vertices) < 2 {
				log.Printf("not enough vertices in graph %s to perform queries", graphName)
				continue
			}

			var totalDijkstraTime time.Duration
			var totalCCHQueryTime time.Duration
			mismatches := 0

			for i := 0; i < numQueries; i++ {
				source, target := selectRandomNodes(vertices)

				// Run Dijkstra on original graph
				start := time.Now()
				_, dijkstraDist, _, err := pathfinding.DijkstraShortestPath(originalNetwork.Network, source, target, math.Inf(1))
				dijkstraTime := time.Since(start)
				if err != nil {
					log.Printf("Dijkstra failed for %v to %v on %s: %v", source, target, graphName, err)
					continue
				}
				totalDijkstraTime += dijkstraTime

				// Run CCH Query
				start = time.Now()
				_, cchDist, _, err := cchInstance.Query(source, target)
				cchTime := time.Since(start)
				if err != nil {
					log.Printf("CCH query failed for %v to %v on %s: %v", source, target, graphName, err)
					continue
				}
				totalCCHQueryTime += cchTime

				// Compare distances
				if math.Abs(dijkstraDist-cchDist) > 1e-6 { // Using a tolerance for float comparison
					log.Printf("Distance mismatch for %v to %v on %s: Dijkstra=%.2f, CCH=%.2f", source, target, graphName, dijkstraDist, cchDist)
					mismatches++
				}
			}

			result := CCHQueryExperimentResult{
				GraphName:       graphName,
				AvgDijkstraTime: totalDijkstraTime / time.Duration(numQueries),
				AvgCCHQueryTime: totalCCHQueryTime / time.Duration(numQueries),
				Mismatches:      mismatches,
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

	headers := []string{"Graph", "AvgDijkstraTime(ms)", "AvgCCHQueryTime(ms)", "Mismatches"}
	writer.Write(headers)

	for _, result := range results {
		row := []string{
			result.GraphName,
			fmt.Sprintf("%.3f", float64(result.AvgDijkstraTime.Nanoseconds())/1e6),
			fmt.Sprintf("%.3f", float64(result.AvgCCHQueryTime.Nanoseconds())/1e6),
			strconv.Itoa(result.Mismatches),
		}
		writer.Write(row)
	}

	log.Printf("CCH query experiment results written to %s", resultsPath)
}
