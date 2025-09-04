# Efficient Route Planning

<p align="center">
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT">
  <img src="https://img.shields.io/badge/Go-1.20-blue.svg" alt="Go">
  <img src="https://img.shields.io/badge/Vue.js-3.x-brightgreen.svg" alt="Vue.js">
  <img src="https://img.shields.io/badge/deck.gl-8.x-blueviolet.svg" alt="deck.gl">
  <img src="https://img.shields.io/badge/MapLibre%20GL%20JS-2.x-lightgrey.svg" alt="MapLibre GL JS">
</p>

<p align="center">
  <img src="./efficient route planning.gif" alt="Project Demo GIF">
</p>

This project provides an interactive visualization platform for comparing pathfinding algorithms on real road networks. It implements and visualizes Contraction Hierarchies (CH), Customizable Contraction Hierarchies (CCH) and Dijkstra's algorithm.

## Features

*   **Interactive Map Visualization:** WebGL-based rendering with `deck.gl` and `Vue.js` for dynamic map interactions.
*   **Algorithm Comparison:** Visualize and compare pathfinding results from Dijkstra, CH, and CCH.
*   **Hierarchical Structure Visualization:** Three-dimensional arc rendering to distinguish shortcut edges and hierarchical levels.
*   **Geocoding Integration:** Search for locations by address or city and find the nearest graph vertex.
*   **Performance Metrics:** Display real-time performance metrics for various algorithms.

## Technologies Used

*   **Backend:** Go (for pathfinding algorithms and API)
*   **Frontend:** Vue.js, deck.gl, MapLibre GL JS
*   **Data:** OpenStreetMap road networks, KaHIP for graph partitioning

## Getting Started

To set up and run the project locally:

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/efficient-routeplanning.git
    cd efficient-routeplanning
    ```
2.  **Backend Setup (Go):**
    ```bash
    go mod tidy
    go build ./cmd/efficient-routeplanning
    # Run the backend server
    ./efficient-routeplanning
    ```
3.  **Frontend Setup (Vue.js):**
    ```bash
    cd frontend
    npm install
    npm run dev
    ```
    Open your browser to `http://localhost:5173` (or the address shown in your terminal).

## Experiments and Results

The `experiments/` directory contains Go programs for benchmarking the implemented algorithms. Results are stored in the `results/` directory, including CSV data and generated plots. Refer to `experiments/README.md` for more details on running experiments.
A written report that summarizes my findings is found in `report.pdf`

## Demonstration

Two short videos demonstrating the tool are available in `demo1.mov` and `demo2.mov`.  

- **Demo 1** visualizes the user taking a look at CH shortcuts as three-dimensional arcs, highlighting how CH computes its Graph.
- **Demo 2** demonstrates the geocoding functionality, showing a query from "Denkendorf" to "Stetten" in Baden-Württemberg and mapping the result onto the road network for accurate pathfinding.