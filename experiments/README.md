# Route Planning Experiments

This directory contains a suite of experiments to evaluate the performance of different route planning algorithms, specifically Contraction Hierarchies (CH) and Customizable Contraction Hierarchies (CCH).

## How to Run Experiments

All experiments can be run via the main executable located in `cmd/ch_experiment/main.go`. You can specify which experiment to run using the `--experiment` flag.

To run a specific experiment (e.g., `ch`):
```bash
go run cmd/ch_experiment/main.go --experiment ch
```

A convenience script is provided to run all experiments sequentially:
```bash
./run_all_experiments.sh
```

## Experiments

Below is a description of each available experiment.

### 1. Contraction Hierarchies (CH) - Preprocessing

- **Flag**: `ch`
- **Description**: This experiment runs the standard Contraction Hierarchies preprocessing on all road networks (`osm*.txt` files) found in `data/RoadNetworks`.
- **Metrics Measured**:
    - Preprocessing time.
    - Number of shortcuts added.
- **Output Files**:
    - `ch_experiment_results.csv`: A CSV file containing the measured metrics for each graph.
    - `data/preprocessed/ch_*.gob`: Preprocessed CH graph data for each road network, saved in gob format.

### 2. Contraction Hierarchies (CH) - Query

- **Flag**: `query`
- **Description**: This experiment evaluates the query performance of the preprocessed CH graphs. For each graph, it selects 100 random source-target pairs and compares the results of a standard Dijkstra search on the original graph against the CH-accelerated query.
- **Metrics Measured**:
    - Average query time for standard Dijkstra.
    - Average query time for CH Dijkstra.
    - Average number of nodes popped from the priority queue for both algorithms.
    - Correctness check to ensure path distances are identical.
- **Output Files**:
    - `query_experiment_results.csv`: A CSV file containing the measured performance metrics.

### 3. Customizable Contraction Hierarchies (CCH) - Preprocessing

- **Flag**: `cch_preprocess`
- **Description**: This experiment runs the metric-independent preprocessing phase for CCH. It uses the node orderings provided in the `data/KaHIP` directory.
- **Metrics Measured**:
    - Preprocessing time.
    - Number of shortcuts added.
    - Average and maximum number of triangles considered per contracted node.
- **Output Files**:
    - `cch_preprocess_experiment_results.csv`: A CSV file with the preprocessing metrics.
    - `data/preprocessed/cch_*.gob`: Preprocessed CCH graph data (metric-independent) for each road network.

### 4. Customizable Contraction Hierarchies (CCH) - Customization

- **Flag**: `cch_customization`
- **Description**: This experiment focuses on the metric-dependent customization phase of CCH. It runs the customization twice: once with the original edge weights from the graph files and five times with random edge weights.
- **Metrics Measured**:
    - Customization time with original weights.
    - Average customization time over the five runs with random weights.
- **Output Files**:
    - `cch_customization_experiment_results.csv`: A CSV file comparing the customization times.

### 5. Customizable Contraction Hierarchies (CCH) - Query

- **Flag**: `cch_query`
- **Description**: This experiment evaluates the query performance of a fully customized CCH graph. After customizing the CCH with the original edge weights, it selects 100 random source-target pairs and compares the CCH query results against a standard Dijkstra search on the original graph.
- **Metrics Measured**:
    - Average query time for standard Dijkstra.
    - Average query time for a CCH query.
    - Correctness check to ensure path distances are identical.
- **Output Files**:
    - `cch_query_experiment_results.csv`: A CSV file with the query performance metrics.
