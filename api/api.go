package api

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/PaulMue0/efficient-routeplanning/internal/cch"
	"github.com/PaulMue0/efficient-routeplanning/internal/ch"
	parser "github.com/PaulMue0/efficient-routeplanning/internal/parser"
	pathfinding "github.com/PaulMue0/efficient-routeplanning/internal/pathfinding"
	preprocessed_graph "github.com/PaulMue0/efficient-routeplanning/internal/preprocessed_graph"
	graph "github.com/PaulMue0/efficient-routeplanning/pkg/collection/graph"
)

var (
	cchInstance     *cch.CCH
	chInstance      *ch.ContractionHierarchies
	cchNetwork      *graph.Graph
	originalWeights map[edgeKey]int // Store original edge weights
	mu              sync.RWMutex
)

type edgeKey struct {
	from graph.VertexId
	to   graph.VertexId
}

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func graphHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	mu.RLock()
	defer mu.RUnlock()

	if cchNetwork == nil {
		http.Error(w, "Graph not initialized", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(cchNetwork); err != nil {
		http.Error(w, "Failed to encode graph", http.StatusInternalServerError)
		log.Printf("Error encoding graph: %v", err)
	}
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

	// Store original edge weights
	originalWeights = make(map[edgeKey]int)
	for from, edges := range cchNetwork.Edges {
		for to, edge := range edges {
			originalWeights[edgeKey{from, to}] = edge.Weight
		}
	}

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

	// Apply CORS middleware to all handlers
	http.HandleFunc("/api/cch", corsMiddleware(cchHandler))
	http.HandleFunc("/api/ch", corsMiddleware(chHandler))
	http.HandleFunc("/api/cch/query", corsMiddleware(cchQueryHandler))
	http.HandleFunc("/api/ch/query", corsMiddleware(chQueryHandler))
	http.HandleFunc("/api/ch/query/nounpack", corsMiddleware(chQueryNoUnpackHandler))
	http.HandleFunc("/api/cch/update", corsMiddleware(cchUpdateHandler))
	http.HandleFunc("/api/graph", corsMiddleware(graphHandler))
	http.HandleFunc("/api/dijkstra/query", corsMiddleware(dijkstraQueryHandler))

	log.Println("Starting API server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func cchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if cchInstance == nil {
		http.Error(w, "CCH not initialized", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(cchInstance); err != nil {
		http.Error(w, "Failed to encode CCH instance", http.StatusInternalServerError)
		log.Printf("Error encoding CCH: %v", err)
	}
}

func chHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if chInstance == nil {
		http.Error(w, "CH not initialized", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(chInstance); err != nil {
		http.Error(w, "Failed to encode CH instance", http.StatusInternalServerError)
		log.Printf("Error encoding CH: %v", err)
	}
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

	mu.RLock()
	defer mu.RUnlock()

	if cchNetwork == nil {
		log.Println("cchNetwork is nil")
		http.Error(w, "Graph not initialized", http.StatusInternalServerError)
		return
	}

	start := time.Now()
	path, weight, _, err := pathfinding.DijkstraShortestPath(cchNetwork, graph.VertexId(from), graph.VertexId(to), math.MaxFloat64)
	duration := time.Since(start)
	queryTimeMs := float64(duration.Nanoseconds()) / 1e6

	if err != nil {
		http.Error(w, "Query failed: no path found", http.StatusNotFound)
		log.Printf("Dijkstra query failed: %v", err)
		return
	}

	var pathEdges []PathEdge
	for i := 0; i < len(path)-1; i++ {
		u := path[i]
		v := path[i+1]
		edge, ok := cchNetwork.Edges[u][v]
		if ok {
			pathEdges = append(pathEdges, PathEdge{From: u, To: v, Weight: float64(edge.Weight), IsShortcut: false})
		} else {
			log.Printf("Warning: No edge found between %d and %d in cchNetwork for Dijkstra", u, v)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(QueryResponse{Path: pathEdges, Weight: weight, QueryTimeMs: queryTimeMs})
}

type PathEdge struct {
	From       graph.VertexId `json:"From"`
	To         graph.VertexId `json:"To"`
	Weight     float64        `json:"Weight"`
	IsShortcut bool           `json:"IsShortcut"`
}

type QueryResponse struct {
	Path        []PathEdge `json:"path"`
	Weight      float64    `json:"weight"`
	QueryTimeMs float64    `json:"queryTimeMs"`
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

	start := time.Now()
	path, weight, _, err := cchInstance.Query(graph.VertexId(from), graph.VertexId(to))
	duration := time.Since(start)
	queryTimeMs := float64(duration.Nanoseconds()) / 1e6

	if err != nil {
		http.Error(w, "Query failed: no path found", http.StatusNotFound)
		log.Printf("CCH query failed: %v", err)
		return
	}

	var pathEdges []PathEdge
	for i := 0; i < len(path)-1; i++ {
		u := path[i]
		v := path[i+1]

		edge, ok := cchInstance.UpwardsGraph.Edges[u][v]
		if !ok {
			edge, ok = cchInstance.DownwardsGraph.Edges[u][v]
		}

		if ok {
			pathEdges = append(pathEdges, PathEdge{From: u, To: v, Weight: float64(edge.Weight), IsShortcut: edge.IsShortcut})
		} else {
			log.Printf("Warning: No edge found between %d and %d in CCH graphs for unpacked path", u, v)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(QueryResponse{Path: pathEdges, Weight: weight, QueryTimeMs: queryTimeMs})
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

	if chInstance == nil {
		http.Error(w, "CH not initialized", http.StatusInternalServerError)
		return
	}

	start := time.Now()
	path, weight, _, err := chInstance.Query(graph.VertexId(from), graph.VertexId(to))
	duration := time.Since(start)
	queryTimeMs := float64(duration.Nanoseconds()) / 1e6

	if err != nil {
		http.Error(w, "Query failed: no path found", http.StatusNotFound)
		log.Printf("CH query failed: %v", err)
		return
	}

	var pathEdges []PathEdge
	for i := 0; i < len(path)-1; i++ {
		u := path[i]
		v := path[i+1]

		edge, ok := chInstance.UpwardsGraph.Edges[u][v]
		if !ok {
			edge, ok = chInstance.DownwardsGraph.Edges[u][v]
		}

		if ok {
			pathEdges = append(pathEdges, PathEdge{From: u, To: v, Weight: float64(edge.Weight), IsShortcut: edge.IsShortcut})
		} else {
			log.Printf("Warning: No edge found between %d and %d in CH graphs for unpacked path", u, v)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(QueryResponse{Path: pathEdges, Weight: weight, QueryTimeMs: queryTimeMs})
}

func chQueryNoUnpackHandler(w http.ResponseWriter, r *http.Request) {
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

	if chInstance == nil {
		http.Error(w, "CH not initialized", http.StatusInternalServerError)
		return
	}

	start := time.Now()
	path, weight, _, err := chInstance.QueryNoUnpack(graph.VertexId(from), graph.VertexId(to))
	duration := time.Since(start)
	queryTimeMs := float64(duration.Nanoseconds()) / 1e6

	if err != nil {
		http.Error(w, "Query failed: no path found", http.StatusNotFound)
		log.Printf("CH query failed: %v", err)
		return
	}

	var pathEdges []PathEdge
	for i := 0; i < len(path)-1; i++ {
		u := path[i]
		v := path[i+1]

		edge, ok := chInstance.UpwardsGraph.Edges[u][v]
		if !ok {
			edge, ok = chInstance.DownwardsGraph.Edges[u][v]
		}

		if ok {
			pathEdges = append(pathEdges, PathEdge{From: u, To: v, Weight: float64(edge.Weight), IsShortcut: edge.IsShortcut})
		} else {
			log.Printf("Warning: No edge found between %d and %d in CH graphs", u, v)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(QueryResponse{Path: pathEdges, Weight: weight, QueryTimeMs: queryTimeMs})
}

type EdgeUpdate struct {
	From   graph.VertexId `json:"from"`
	To     graph.VertexId `json:"to"`
	Weight string         `json:"weight"`
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

	mu.Lock()
	defer mu.Unlock()

	if cchNetwork == nil {
		http.Error(w, "Graph not initialized", http.StatusInternalServerError)
		return
	}

	for _, update := range updates {
		var actualWeight int

		// Handle blocking/unblocking edges
		if update.Weight == "inf" {
			actualWeight = math.MaxInt32 // Use MaxInt32 to avoid overflow issues
		} else if update.Weight == "restore" {
			// Restore original weight
			key := edgeKey{from: update.From, to: update.To}
			if origWeight, exists := originalWeights[key]; exists {
				actualWeight = origWeight
			} else {
				log.Printf("Original weight not found for edge %d->%d, using default weight 1", update.From, update.To)
				actualWeight = 1
			}
		} else {
			parsedWeight, err := strconv.Atoi(update.Weight)
			if err != nil {
				log.Printf("Invalid weight format for edge from %d to %d: %v", update.From, update.To, err)
				continue
			}
			actualWeight = parsedWeight
		}

		// Update both directions (undirected graph)
		err := cchNetwork.UpdateEdge(update.From, update.To, actualWeight, false, 0)
		if err != nil {
			log.Printf("Failed to update edge from %d to %d: %v", update.From, update.To, err)
		}

		err = cchNetwork.UpdateEdge(update.To, update.From, actualWeight, false, 0)
		if err != nil {
			log.Printf("Failed to update edge from %d to %d: %v", update.To, update.From, err)
		}
	}

	if cchInstance == nil {
		http.Error(w, "CCH not initialized", http.StatusInternalServerError)
		return
	}

	// Re-customize CCH with updated weights
	cchInstance.Customize(cchNetwork)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
