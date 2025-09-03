package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PaulMue0/efficient-routeplanning/internal/cch"
	"github.com/PaulMue0/efficient-routeplanning/internal/ch"
	parser "github.com/PaulMue0/efficient-routeplanning/internal/parser"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

var (
	cchInstance *cch.CCH
	chInstance  *ch.ContractionHierarchies
)

func loadAndPreprocess() {
	name := "osm1.txt"
	dataDir := "../../data/RoadNetworks"
	fileSystem := os.DirFS(dataDir)
	network, err := parser.NewNetworkFromFS(fileSystem, name)
	if err != nil {
		log.Fatalf("Failed to load graph: %v", err)
	}
	log.Printf("File: %s, NumNodes: %d, NumEdges: %d", name, network.NumNodes, network.NumEdges)

	// Preprocess CCH
	cchInst := cch.NewCCH()
	log.Println("Starting CCH preprocessing...")
	start := time.Now()
	err = cchInst.Preprocess(network.Network, "../../data/kaHIP/osm1.ordering")
	if err != nil {
		log.Fatalf("CCH preprocessing failed: %v", err)
	}
	duration := time.Since(start)
	log.Printf("Finished CCH preprocessing in %s", duration)
	cchInstance = cchInst

	cchInst.Customize(network.Network)

	// Preprocess CH
	// Need to reload the network because preprocessing modifies the graph
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

func StartApi() {
	loadAndPreprocess()

	http.HandleFunc("/api/cch", cchHandler)
	http.HandleFunc("/api/ch", chHandler)
	http.HandleFunc("/api/cch/query", cchQueryHandler)
	http.HandleFunc("/api/ch/query", chQueryHandler)

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
