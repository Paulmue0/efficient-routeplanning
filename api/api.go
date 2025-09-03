package api

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PaulMue0/efficient-routeplanning/internal/cch"
	"github.com/PaulMue0/efficient-routeplanning/internal/ch"
	parser "github.com/PaulMue0/efficient-routeplanning/internal/parser"
	pathfinding "github.com/PaulMue0/efficient-routeplanning/internal/pathfinding"
	preprocessed_graph "github.com/PaulMue0/efficient-routeplanning/internal/preprocessed_graph"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

var (
	cchInstance *cch.CCH
	chInstance  *ch.ContractionHierarchies
	cchNetwork  *graph.Graph
)

func graphHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cchNetwork)
}

func loadAndPreprocess() {
	name := "osm5.txt"
	dataDir := "../../data/RoadNetworks"
	fileSystem := os.DirFS(dataDir)
	network, err := parser.NewNetworkFromFS(fileSystem, name)
	if err != nil {
		log.Fatalf("Failed to load graph: %v", err)
	}
	log.Printf("File: %s, NumNodes: %d, NumEdges: %d", name, network.NumNodes, network.NumEdges)

	cchNetwork = network.Network

	// Preprocess CCH
	cchInst := cch.NewCCH()
	log.Println("Starting CCH preprocessing...")
	start := time.Now()
	err = cchInst.Preprocess(cchNetwork, "../../data/kaHIP/osm5.ordering")
	if err != nil {
		log.Fatalf("CCH preprocessing failed: %v", err)
	}
	duration := time.Since(start)
	log.Printf("Finished CCH preprocessing in %s", duration)
	cchInstance = cchInst

	cchInst.Customize(cchNetwork)

	// Preprocess CH
	// Need to reload the network because preprocessing modifies the graph
	// Try to load preprocessed CH from file
	chFilePath := "../../data/preprocessed/ch_osm7.gob"
	log.Printf("Attempting to load preprocessed CH from %s", chFilePath)
	chFile, err := preprocessed_graph.ReadCHFile(chFilePath)
	if err == nil {
		log.Printf("Successfully loaded preprocessed CH from %s", chFilePath)
		chInstance = chFile.ToCH()
	} else {
		log.Printf("Failed to load preprocessed CH (%v), performing preprocessing instead.", err)
		network, err = parser.NewNetworkFromFS(fileSystem, name)
		if err != nil {
			log.Fatalf("Failed to reload graph for CH: %v", err)
		}
		chInst := ch.NewContractionHierarchies()
		log.Println("Starting CH preprocessing...")
		start = time.Now()
		chInst.Preprocess(network.Network)
		duration = time.Since(start)
		log.Printf("Finished CH preprocessing in %s", duration)
		chInstance = chInst
	}
}

func StartApi() {
	loadAndPreprocess()

	http.HandleFunc("/api/cch", cchHandler)
	http.HandleFunc("/api/ch", chHandler)
	http.HandleFunc("/api/cch/query", cchQueryHandler)
	http.HandleFunc("/api/ch/query", chQueryHandler)
	http.HandleFunc("/api/cch/update", cchUpdateHandler)
	http.HandleFunc("/api/graph", graphHandler)
	http.HandleFunc("/api/dijkstra/query", dijkstraQueryHandler)

	log.Println("Starting API server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func cchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cchInstance)
}

func chHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chInstance)
}

func dijkstraQueryHandler(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	log.Printf("Dijkstra query: from=%s, to=%s", fromStr, toStr)

	from, err := strconv.Atoi(fromStr)
	if err != nil {
		http.Error(w, "Invalid 'from' parameter", http.StatusBadRequest)
		return
	}

	to, err := strconv.Atoi(toStr)
	if err != nil {
		http.Error(w, "Invalid 'to' parameter", http.StatusBadRequest)
		return
	}

	if cchNetwork == nil { // Dijkstra operates on the base graph
		log.Println("cchNetwork is nil")
		http.Error(w, "Graph not initialized", http.StatusInternalServerError)
		return
	}

	start := time.Now() // Start timing
	// DijkstraShortestPath expects a graph.Graph, source, target, and a bound.
	// We can use math.MaxFloat64 as the bound for an unbounded search.
	path, weight, _, err := pathfinding.DijkstraShortestPath(cchNetwork, graph.VertexId(from), graph.VertexId(to), math.MaxFloat64)
	duration := time.Since(start)                        // End timing
	queryTimeMs := float64(duration.Nanoseconds()) / 1e6 // Convert to milliseconds

	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		log.Printf("Dijkstra query failed: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(QueryResponse{Path: path, Weight: weight, QueryTimeMs: queryTimeMs})
}

type QueryResponse struct {
	Path        []graph.VertexId `json:"path"`
	Weight      float64          `json:"weight"`
	QueryTimeMs float64          `json:"queryTimeMs"`
}

func cchQueryHandler(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	log.Printf("CCH query: from=%s, to=%s", fromStr, toStr)

	from, err := strconv.Atoi(fromStr)
	if err != nil {
		http.Error(w, "Invalid 'from' parameter", http.StatusBadRequest)
		return
	}

	to, err := strconv.Atoi(toStr)
	if err != nil {
		http.Error(w, "Invalid 'to' parameter", http.StatusBadRequest)
		return
	}

	if cchInstance == nil {
		log.Println("cchInstance is nil")
		http.Error(w, "CCH not initialized", http.StatusInternalServerError)
		return
	}
	log.Printf("cchInstance is not nil, UpwardsGraph has %d vertices", len(cchInstance.UpwardsGraph.Vertices))

	start := time.Now() // Start timing
	path, weight, _, err := cchInstance.Query(graph.VertexId(from), graph.VertexId(to))
	duration := time.Since(start)                        // End timing
	queryTimeMs := float64(duration.Nanoseconds()) / 1e6 // Convert to milliseconds

	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		log.Printf("CCH query failed: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(QueryResponse{Path: path, Weight: weight, QueryTimeMs: queryTimeMs})
}

func chQueryHandler(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := strconv.Atoi(fromStr)
	if err != nil {
		http.Error(w, "Invalid 'from' parameter", http.StatusBadRequest)
		return
	}

	to, err := strconv.Atoi(toStr)
	if err != nil {
		http.Error(w, "Invalid 'to' parameter", http.StatusBadRequest)
		return
	}

	start := time.Now() // Start timing
	path, weight, _, err := chInstance.Query(graph.VertexId(from), graph.VertexId(to))
	duration := time.Since(start)                        // End timing
	queryTimeMs := float64(duration.Nanoseconds()) / 1e6 // Convert to milliseconds

	if err != nil {
		http.Error(w, "Query failed", http.StatusInternalServerError)
		log.Printf("CH query failed: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(QueryResponse{Path: path, Weight: weight, QueryTimeMs: queryTimeMs})
}

type EdgeUpdate struct {
	From   graph.VertexId `json:"from"`
	To     graph.VertexId `json:"to"`
	Weight string         `json:"weight"` // Changed to string
}

func cchUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		log.Printf("Error reading request body: %v", err)
		return
	}
	log.Printf("Received CCH update request body: %s", string(body))

	var updates []EdgeUpdate
	if err := json.Unmarshal(body, &updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", err)
		return
	}

	if cchNetwork == nil {
		http.Error(w, "Graph not initialized", http.StatusInternalServerError)
		return
	}

	for _, update := range updates {
		var actualWeight int
		if update.Weight == "inf" {
			actualWeight = math.MaxInt
		} else {
			parsedWeight, err := strconv.Atoi(update.Weight)
			if err != nil {
				log.Printf("Invalid weight format for edge from %d to %d: %v", update.From, update.To, err)
				continue // Skip this update and continue with others
			}
			actualWeight = parsedWeight
		}

		err := cchNetwork.UpdateEdge(update.To, update.From, actualWeight, false, 0)
		if err != nil {
			log.Printf("Failed to update edge from %d to %d: %v", update.From, update.To, err)
		}

		err = cchNetwork.UpdateEdge(update.From, update.To, actualWeight, false, 0)
		if err != nil {
			log.Printf("Failed to update edge from %d to %d: %v", update.From, update.To, err)
		}
	}

	if cchInstance == nil {
		http.Error(w, "CCH not initialized", http.StatusInternalServerError)
		return
	}

	cchInstance.Customize(cchNetwork)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cchInstance)
}
