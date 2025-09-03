package experiments

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
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

type QueryExperimentResult struct {
	GraphName            string
	AvgDijkstraTime      time.Duration
	AvgCHDijkstraTime    time.Duration
	AvgDijkstraNodesPopped int
	AvgCHNodesPopped     int
	Mismatches           int
}

func RunQueryExperiment() {
	preprocessedDir := "./data/preprocessed"
	roadNetworksDir := "./data/RoadNetworks"
	resultsPath := "./query_experiment_results.csv"
	numQueries := 100

	files, err := os.ReadDir(preprocessedDir)
	if err != nil {
		log.Fatalf("failed to read preprocessed directory: %v", err)
	}

	var results []QueryExperimentResult

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "ch_osm") && strings.HasSuffix(file.Name(), ".gob") {
			graphName := strings.TrimSuffix(strings.TrimPrefix(file.Name(), "ch_"), ".gob") + ".txt"
			log.Printf("Processing graph: %s", graphName)

			// Load original graph
			fileSystem := os.DirFS(roadNetworksDir)
			originalNetwork, err := parser.NewNetworkFromFS(fileSystem, graphName)
			if err != nil {
				log.Printf("failed to load original graph %s: %v", graphName, err)
				continue
			}

			// Load CH graph
			chFilePath := filepath.Join(preprocessedDir, file.Name())
			preprocessedFile, err := preprocessed_graph.ReadCHFile(chFilePath)
			if err != nil {
				log.Printf("failed to read preprocessed graph for %s: %v", graphName, err)
				continue
			}
			chInstance := preprocessedFile.ToCH()

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
			var totalCHDijkstraTime time.Duration
			var totalDijkstraNodesPopped int
			var totalCHNodesPopped int
			mismatches := 0

			for i := 0; i < numQueries; i++ {
				source, target := selectRandomNodes(vertices)

				// Run Dijkstra on original graph
				start := time.Now()
				_, dijkstraDist, dijkstraNodesPopped, err := pathfinding.DijkstraShortestPath(originalNetwork.Network, source, target, math.Inf(1))
				dijkstraTime := time.Since(start)
				if err != nil {
					log.Printf("Dijkstra failed for %s to %s on %s: %v", source, target, graphName, err)
					continue
				}
				totalDijkstraTime += dijkstraTime
				totalDijkstraNodesPopped += dijkstraNodesPopped

				// Run CH-Dijkstra on CH graph
				start = time.Now()
				_, chDist, chNodesPopped, err := chInstance.Query(source, target)
				chTime := time.Since(start)
				if err != nil {
					log.Printf("CH-Dijkstra failed for %s to %s on %s: %v", source, target, graphName, err)
					continue
				}
				totalCHDijkstraTime += chTime
				totalCHNodesPopped += chNodesPopped

				// Compare distances
				if math.Abs(dijkstraDist-chDist) > 1e-6 { // Using a tolerance for float comparison
					log.Printf("Distance mismatch for %s to %s on %s: Dijkstra=%.2f, CH=%.2f", source, target, graphName, dijkstraDist, chDist)
					mismatches++
				}
			}

			result := QueryExperimentResult{
				GraphName:            graphName,
				AvgDijkstraTime:      totalDijkstraTime / time.Duration(numQueries),
				AvgCHDijkstraTime:    totalCHDijkstraTime / time.Duration(numQueries),
				AvgDijkstraNodesPopped: totalDijkstraNodesPopped / numQueries,
				AvgCHNodesPopped:     totalCHNodesPopped / numQueries,
				Mismatches:           mismatches,
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

	headers := []string{"Graph", "AvgDijkstraTime(ms)", "AvgCHDijkstraTime(ms)", "AvgDijkstraNodesPopped", "AvgCHNodesPopped", "Mismatches"}
	writer.Write(headers)

	for _, result := range results {
		row := []string{
			result.GraphName,
			fmt.Sprintf("%.3f", float64(result.AvgDijkstraTime.Nanoseconds())/1e6),
			fmt.Sprintf("%.3f", float64(result.AvgCHDijkstraTime.Nanoseconds())/1e6),
			strconv.Itoa(result.AvgDijkstraNodesPopped),
			strconv.Itoa(result.AvgCHNodesPopped),
			strconv.Itoa(result.Mismatches),
		}
		writer.Write(row)
	}

	log.Printf("Query experiment results written to %s", resultsPath)
}

func selectRandomNodes(nodes []graph.VertexId) (graph.VertexId, graph.VertexId) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sourceIndex := r.Intn(len(nodes))
	targetIndex := r.Intn(len(nodes))
	for sourceIndex == targetIndex {
		targetIndex = r.Intn(len(nodes))
	}
	return nodes[sourceIndex], nodes[targetIndex]
}
