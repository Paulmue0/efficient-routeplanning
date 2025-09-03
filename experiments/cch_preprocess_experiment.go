package experiments

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PaulMue0/efficient-routeplanning/internal/cch"
	"github.com/PaulMue0/efficient-routeplanning/internal/parser"
	"github.com/PaulMue0/efficient-routeplanning/internal/preprocessed_graph"
)

type CCHPreprocessExperimentResult struct {
	GraphName         string
	PreprocessingTime time.Duration
	ShortcutsAdded    int
	AvgTriangles      float64
	MaxTriangles      int
}

func RunCCHPreprocessExperiment() {
	dataDir := "./data/RoadNetworks"
	orderingDir := "./data/KaHIP"
	resultsPath := "./cch_preprocess_experiment_results.csv"
	preprocessedPath := "./data/preprocessed"

	if _, err := os.Stat(preprocessedPath); os.IsNotExist(err) {
		os.MkdirAll(preprocessedPath, 0755)
	}

	files, err := os.ReadDir(dataDir)
	if err != nil {
		log.Fatalf("failed to read data directory: %v", err)
	}

	var results []CCHPreprocessExperimentResult

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "osm") && strings.HasSuffix(file.Name(), ".txt") {
			graphName := file.Name()
			log.Printf("Processing graph: %s", graphName)

			// Load graph
			fileSystem := os.DirFS(dataDir)
			network, err := parser.NewNetworkFromFS(fileSystem, graphName)
			if err != nil {
				log.Printf("failed to load graph %s: %v", graphName, err)
				continue
			}

			// Get ordering file path
			orderingFile := strings.TrimSuffix(graphName, ".txt") + ".ordering"
			orderingFilePath := filepath.Join(orderingDir, orderingFile)
			if _, err := os.Stat(orderingFilePath); os.IsNotExist(err) {
				log.Printf("ordering file not found for %s, skipping", graphName)
				continue
			}

			// Run CCH preprocessing
			cchInstance := cch.NewCCH()
			start := time.Now()
			err = cchInstance.Preprocess(network.Network, orderingFilePath)
			if err != nil {
				log.Printf("CCH preprocessing failed for %s: %v", graphName, err)
				continue
			}
			duration := time.Since(start)

			avgTriangles := 0.0
			if len(cchInstance.ContractionOrder) > 0 {
				avgTriangles = float64(cchInstance.TotalTriangles) / float64(len(cchInstance.ContractionOrder))
			}

			result := CCHPreprocessExperimentResult{
				GraphName:         graphName,
				PreprocessingTime: duration,
				ShortcutsAdded:    cchInstance.ShortcutsAdded,
				AvgTriangles:      avgTriangles,
				MaxTriangles:      cchInstance.MaxTriangles,
			}
			results = append(results, result)

			// Save preprocessed graph
			preprocessedFile := preprocessed_graph.FromCCH(cchInstance)
			outputPath := filepath.Join(preprocessedPath, "cch_"+strings.TrimSuffix(graphName, ".txt")+".gob")
			err = preprocessedFile.Write(outputPath)
			if err != nil {
				log.Printf("failed to write preprocessed cch graph for %s: %v", graphName, err)
			}

			log.Printf("Finished processing %s in %s, shortcuts added: %d", graphName, duration, cchInstance.ShortcutsAdded)
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

	headers := []string{"Graph", "PreprocessingTime(ms)", "ShortcutsAdded", "AvgTriangles", "MaxTriangles"}
	writer.Write(headers)

	for _, result := range results {
		row := []string{
			result.GraphName,
			strconv.FormatInt(result.PreprocessingTime.Milliseconds(), 10),
			strconv.Itoa(result.ShortcutsAdded),
			fmt.Sprintf("%.2f", result.AvgTriangles),
			strconv.Itoa(result.MaxTriangles),
		}
		writer.Write(row)
	}

	log.Printf("CCH preprocess experiment results written to %s", resultsPath)
}
