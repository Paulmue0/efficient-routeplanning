
package experiments

import (
	"encoding/csv"
	
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PaulMue0/efficient-routeplanning/internal/ch"
	"github.com/PaulMue0/efficient-routeplanning/internal/parser"
	"github.com/PaulMue0/efficient-routeplanning/internal/preprocessed_graph"
)

type CHExperimentResult struct {
	GraphName         string
	PreprocessingTime time.Duration
	ShortcutsAdded    int
}

func RunCHExperiment() {
	dataDir := "./data/RoadNetworks"
	resultsPath := "./ch_experiment_results.csv"
	preprocessedPath := "./data/preprocessed"

	if _, err := os.Stat(preprocessedPath); os.IsNotExist(err) {
		os.MkdirAll(preprocessedPath, 0755)
	}

	files, err := os.ReadDir(dataDir)
	if err != nil {
		log.Fatalf("failed to read data directory: %v", err)
	}

	var results []CHExperimentResult

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "osm") && strings.HasSuffix(file.Name(), ".txt") {
			graphName := file.Name()
			log.Printf("Processing graph: %s", graphName)

			fileSystem := os.DirFS(filepath.Join(dataDir))
			network, err := parser.NewNetworkFromFS(fileSystem, graphName)
			if err != nil {
				log.Printf("failed to load graph %s: %v", graphName, err)
				continue
			}

			ch.ShortcutsAdded = 0 // Reset shortcut counter
			chInstance := ch.NewContractionHierarchies()

			start := time.Now()
			chInstance.Preprocess(network.Network)
			duration := time.Since(start)

			result := CHExperimentResult{
				GraphName:         graphName,
				PreprocessingTime: duration,
				ShortcutsAdded:    ch.ShortcutsAdded,
			}
			results = append(results, result)

			// Save preprocessed graph
			preprocessedFile := preprocessed_graph.FromCH(chInstance)
			outputPath := filepath.Join(preprocessedPath, "ch_"+strings.TrimSuffix(graphName, ".txt")+".gob")
			err = preprocessedFile.WriteCH(outputPath)
			if err != nil {
				log.Printf("failed to write preprocessed graph for %s: %v", graphName, err)
			}

			log.Printf("Finished processing %s in %s, shortcuts added: %d", graphName, duration, ch.ShortcutsAdded)
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

	headers := []string{"Graph", "PreprocessingTime(ms)", "ShortcutsAdded"}
	writer.Write(headers)

	for _, result := range results {
		row := []string{
			result.GraphName,
			strconv.FormatInt(result.PreprocessingTime.Milliseconds(), 10),
			strconv.Itoa(result.ShortcutsAdded),
		}
		writer.Write(row)
	}

	log.Printf("Experiment results written to %s", resultsPath)
}
