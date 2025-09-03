<script setup>
import { ref, onMounted, watch } from 'vue';
import GraphVisualization from './components/GraphVisualization.vue';
import RouteSelector from './components/RouteSelector.vue';
import { updateEdgeWeights } from './services/edgeService';

const geoJsonData = ref(null);
const shortcutsGeoJsonData = ref(null);
const baseGeoJsonData = ref(null);
const startNode = ref(null);
const endNode = ref(null);
const shortestPath = ref(null);
const pathShortcuts = ref(null); // New ref for path shortcuts to be shown as arcs
const pathWeight = ref(null);
const queryTimeMs = ref(null);
const selectedAlgorithm = ref('ch'); // Default to CH
const verticesMap = ref(new Map());
const blockedEdges = ref([]); // New ref to store blocked edges
const showShortcuts = ref(true); // New ref to control shortcut visibility

const viewState = ref({
  longitude: 9.244557,
  latitude: 48.667421,
  zoom: 11,
  pitch: 0,
  bearing: 0
});

const handleViewStateChange = (newViewState) => {
  viewState.value = newViewState;
};

const handleLayerClick = (info) => {
  if (info.object && info.object.properties.type === 'vertex') {
    const vertexId = info.object.properties.id;

    if (!startNode.value) {
      // First click: set start node
      startNode.value = vertexId;
    } else if (!endNode.value && vertexId !== startNode.value) {
      // Second click: set end node, if different from start
      endNode.value = vertexId;
    } else {
      // Third click or clicking start node again: reset and set new start node
      startNode.value = vertexId;
      endNode.value = null;
      shortestPath.value = null; // Clear previous path
      pathShortcuts.value = null;
      pathWeight.value = null;
      queryTimeMs.value = null;
    }
  }
};

async function handleEdgeClick({ from, to }) {
  if (selectedAlgorithm.value !== 'cch') {
    alert('Edge blocking is only available for CCH algorithm.');
    return;
  }

  const index = blockedEdges.value.findIndex(edge => edge.from === from && edge.to === to);
  let weight = "inf"; // Default to blocking

  if (index !== -1) {
    // Edge is already blocked, unblock it
    blockedEdges.value.splice(index, 1);
    // TODO: Get original weight. For now, we'll just remove it from blocked list.
    // A more robust solution would involve storing original weights or re-fetching the graph.
    weight = "1"; // Assuming a default unblocked weight for now, or fetch original
  } else {
    // Edge is not blocked, block it
    blockedEdges.value.push({ from, to });
  }

  try {
    await updateEdgeWeights([{ from: from, to: to, weight: weight }]);
    // Re-fetch graph data to update visualization and ensure pathfinding uses new weights
    await fetchGraphData(selectedAlgorithm.value);
    // Recalculate shortest path if nodes are selected
    if (startNode.value && endNode.value) {
      await calculateShortestPath();
    }
  } catch (error) {
    console.error('Failed to update edge or re-fetch graph:', error);
    // Revert UI change if backend update fails
    if (index === -1) {
      blockedEdges.value.pop();
    } else {
      blockedEdges.value.splice(index, 0, { from, to });
    }
  }
}

async function calculateShortestPath() {
  if (!startNode.value || !endNode.value) {
    shortestPath.value = null;
    pathShortcuts.value = null;
    pathWeight.value = null;
    queryTimeMs.value = null;
    return;
  }
  try {
    const apiUrl = `/api/${selectedAlgorithm.value.startsWith('ch-nounpack') ? 'ch/query/nounpack' : selectedAlgorithm.value + '/query'}?from=${startNode.value}&to=${endNode.value}`;
    console.log('API URL:', apiUrl, 'Algorithm:', selectedAlgorithm.value);
    const response = await fetch(apiUrl);
    if (!response.ok) throw new Error('Failed to fetch shortest path');
    const result = await response.json();
    console.log('Query result:', result);
    if (result.path && result.path.length > 0) {
      const pathFeatures = [];
      const shortcutFeatures = [];
      
      for (const edge of result.path) {
        console.log('Processing edge:', edge);
        const fromCoords = verticesMap.value.get(edge.From);
        const toCoords = verticesMap.value.get(edge.To);
        console.log('Coordinates:', { from: fromCoords, to: toCoords });
        if (fromCoords && toCoords) {
          if (edge.IsShortcut) {
            console.log('Adding shortcut feature');
            // Add shortcuts as arc features
            shortcutFeatures.push({
              type: 'Feature',
              geometry: { type: 'LineString', coordinates: [fromCoords, toCoords] },
              properties: { IsShortcut: true, weight: edge.Weight, from: edge.From, to: edge.To }
            });
          } else {
            console.log('Adding regular path feature');
            // Add real edges as line features
            pathFeatures.push({
              type: 'Feature',
              geometry: { type: 'LineString', coordinates: [fromCoords, toCoords] },
              properties: { IsShortcut: false, weight: edge.Weight }
            });
          }
        }
      }

      shortestPath.value = pathFeatures.length > 0 ? {
        type: 'FeatureCollection',
        features: pathFeatures
      } : null;
      
      pathShortcuts.value = shortcutFeatures.length > 0 ? {
        type: 'FeatureCollection',
        features: shortcutFeatures
      } : null;
      
      console.log('Final results:', {
        pathFeatures: pathFeatures.length,
        shortcutFeatures: shortcutFeatures.length,
        shortestPath: shortestPath.value,
        pathShortcuts: pathShortcuts.value
      });
      
      pathWeight.value = result.weight;
      queryTimeMs.value = result.queryTimeMs;
    } else {
      shortestPath.value = null;
      pathShortcuts.value = null;
      pathWeight.value = null;
      queryTimeMs.value = null;
    }
  } catch (error) {
    console.error(error);
    shortestPath.value = null;
    pathShortcuts.value = null;
    pathWeight.value = null;
    queryTimeMs.value = null;
  }
}

async function fetchGraphData(algorithm) {
  try {
    // Always fetch the base graph
    const baseGraphResponse = await fetch('/api/graph');
    if (!baseGraphResponse.ok) throw new Error(`HTTP error! status: ${baseGraphResponse.status}`);
    const baseGraphDataRaw = await baseGraphResponse.json();

    const baseFeatures = [];
    const baseProcessedVertices = new Set();

    const processBaseGraph = (graph) => {
      if (!graph) return;
      if (graph.Vertices) {
        for (const vertexId in graph.Vertices) {
          const vertex = graph.Vertices[vertexId];
          if (!baseProcessedVertices.has(vertexId)) {
            baseFeatures.push({
              type: 'Feature',
              geometry: { type: 'Point', coordinates: [vertex.Lon, vertex.Lat] },
              properties: { id: vertex.Id, type: 'vertex' }
            });
            baseProcessedVertices.add(vertexId);
          }
        }
      }
      if (graph.Edges && graph.Vertices) {
        for (const sourceId in graph.Edges) {
          const sourceVertex = graph.Vertices[sourceId];
          if (!sourceVertex) continue;
          for (const targetId in graph.Edges[sourceId]) {
            const targetVertex = graph.Vertices[targetId];
            if (!targetVertex) continue;
            baseFeatures.push({
              type: 'Feature',
              geometry: { type: 'LineString', coordinates: [[sourceVertex.Lon, sourceVertex.Lat], [targetVertex.Lon, targetVertex.Lat]] },
              properties: { source: sourceId, target: targetId, type: 'edge', graph: 'base', weight: graph.Edges[sourceId][targetId].Weight }
            });
          }
        }
      }
    };
    processBaseGraph(baseGraphDataRaw);
    baseGeoJsonData.value = { type: 'FeatureCollection', features: baseFeatures };

    // Now fetch algorithm-specific data
    let apiUrl;
    if (algorithm === 'ch-nounpack') {
      apiUrl = '/api/ch'; // Always fetch the CH graph for both CH and CH-NoUnpack
    } else {
      apiUrl = `/api/${algorithm}`;
    }
    let graphData;
    const features = [];
    const shortcutsFeatures = [];
    const processedVertices = new Set(); // This set is for algorithm-specific graph processing

    if (algorithm === 'dijkstra') {
      // For Dijkstra, the algorithm-specific graph is the base graph itself
      graphData = baseGraphDataRaw;
    } else {
      const response = await fetch(apiUrl);
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      graphData = await response.json();
    }

    // Clear existing verticesMap to avoid stale data for algorithm-specific graph
    verticesMap.value.clear();

    // First populate vertices map from base graph to ensure all vertices are available
    if (baseGraphDataRaw && baseGraphDataRaw.Vertices) {
      for (const vertexId in baseGraphDataRaw.Vertices) {
        const vertex = baseGraphDataRaw.Vertices[vertexId];
        verticesMap.value.set(vertex.Id, [vertex.Lon, vertex.Lat]);
      }
    }

    const processAlgorithmGraph = (graph, graphType) => {
      if (!graph) return;
      if (graph.Vertices) {
        for (const vertexId in graph.Vertices) {
          const vertex = graph.Vertices[vertexId];
          // Only add if not already in vertices map (to preserve base graph vertices)
          if (!verticesMap.value.has(vertex.Id)) {
            verticesMap.value.set(vertex.Id, [vertex.Lon, vertex.Lat]);
          }
          if (!processedVertices.has(vertexId)) {
            // Only add vertices if they are not already in the base graph features
            // This is to avoid duplicate vertices if base graph is also processed here
            // For now, we'll just add them, assuming base graph handles primary vertex rendering
            features.push({
              type: 'Feature',
              geometry: { type: 'Point', coordinates: [vertex.Lon, vertex.Lat] },
              properties: { id: vertex.Id, type: 'vertex' }
            });
            processedVertices.add(vertexId);
          }
        }
      }
      if (graph.Edges && graph.Vertices) {
        for (const sourceId in graph.Edges) {
          const sourceVertex = graph.Vertices[sourceId];
          if (!sourceVertex) continue;
          for (const targetId in graph.Edges[sourceId]) {
            const targetVertex = graph.Vertices[targetId];
            if (!targetVertex) continue;
            const edgeFeature = {
              type: 'Feature',
              geometry: { type: 'LineString', coordinates: [[sourceVertex.Lon, sourceVertex.Lat], [targetVertex.Lon, targetVertex.Lat]] },
              properties: { source: sourceId, target: targetId, type: 'edge', graph: graphType, isShortcut: graph.Edges[sourceId][targetId].IsShortcut, weight: graph.Edges[sourceId][targetId].Weight }
            };
            if (edgeFeature.properties.isShortcut) {
              shortcutsFeatures.push(edgeFeature);
            } else {
              features.push(edgeFeature);
            }
          }
        }
      }
    };

    // Process algorithm-specific graph data
    if (algorithm === 'ch') {
      processAlgorithmGraph(graphData.UpwardsGraph, 'upwards');
      processAlgorithmGraph(graphData.DownwardsGraph, 'downwards');
    } else if (algorithm === 'cch') {
      processAlgorithmGraph(graphData.UpwardsGraph, 'upwards');
      processAlgorithmGraph(graphData.DownwardsGraph, 'downwards');
    } else if (algorithm === 'dijkstra') {
      // For Dijkstra, geoJsonData will be the same as baseGeoJsonData
      processAlgorithmGraph(graphData, 'base');
    }

    geoJsonData.value = { type: 'FeatureCollection', features: features };
    shortcutsGeoJsonData.value = { type: 'FeatureCollection', features: shortcutsFeatures };
  } catch (error) {
    console.error('Failed to fetch and parse graph data:', error);
  }
}

watch(endNode, (newValue) => {
  if (newValue) {
    calculateShortestPath();
  }
});

watch(selectedAlgorithm, async (newAlgorithm) => {
  startNode.value = null;
  endNode.value = null;
  shortestPath.value = null;
  pathShortcuts.value = null;
  pathWeight.value = null;
  queryTimeMs.value = null;
  blockedEdges.value = []; // Clear blocked edges when algorithm changes
  await fetchGraphData(newAlgorithm);
});

onMounted(async () => {
  await fetchGraphData(selectedAlgorithm.value);
});
</script>

<template>
  <GraphVisualization :geo-json-data="geoJsonData" :shortest-path="shortestPath" :path-shortcuts="pathShortcuts" :start-node="startNode"
    :end-node="endNode" :vertices-map="verticesMap" :selected-algorithm="selectedAlgorithm" :view-state="viewState"
    :blocked-edges="blockedEdges" :shortcuts-geo-json-data="shortcutsGeoJsonData" :show-shortcuts="showShortcuts"
    :base-geo-json-data="baseGeoJsonData" @update:view-state="handleViewStateChange" @layer-click="handleLayerClick"
    @edge-click="handleEdgeClick" />
  <RouteSelector :geo-json-data="geoJsonData" :start-node="startNode" :end-node="endNode"
    @update:start-node="startNode = $event" @update:end-node="endNode = $event" />
  <div class="info-overlay">
    <div class="algorithm-selector">
      <label for="algorithm-select">Algorithm:</label>
      <select id="algorithm-select" v-model="selectedAlgorithm">
        <option value="ch">Contraction Hierarchies (CH)</option>
        <option value="ch-nounpack">Contraction Hierarchies (CH) - No Unpack</option>
        <option value="cch">Customizable Contraction Hierarchies (CCH)</option>
        <option value="dijkstra">Dijkstra</option>
      </select>
    </div>
    <p v-if="startNode">Start Node: {{ startNode }}</p>
    <p v-if="endNode">End Node: {{ endNode }}</p>
    <p v-if="pathWeight !== null">Path Weight: {{ pathWeight.toFixed(2) }}</p>
    <p v-if="queryTimeMs">Query Time: {{ queryTimeMs.toFixed(2) }} ms</p>
    <p v-if="startNode && !endNode">Click a vertex to select end node.</p>
    <p v-if="!startNode">Click a vertex to select start node.</p>
    <div class="layer-toggle">
      <input type="checkbox" id="show-shortcuts" v-model="showShortcuts">
      <label for="show-shortcuts">Show Shortcuts</label>
    </div>
  </div>
</template>

<style lang="scss">
.info-overlay {
  position: absolute;
  top: 10px;
  left: 10px;
  background: rgba(255, 255, 255, 0.8);
  padding: 10px;
  border-radius: 5px;
  z-index: 1000; // Ensure it's above the map
  color: black;
  font-family: sans-serif;

  p {
    margin: 0 0 5px 0;

    &:last-child {
      margin-bottom: 0;
    }
  }
}

@import 'maplibre-gl/dist/maplibre-gl.css';
</style>
