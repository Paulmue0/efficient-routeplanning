<script setup>
import { ref, onMounted, watch } from 'vue';
import GraphVisualization from './components/GraphVisualization.vue';
import RouteSelector from './components/RouteSelector.vue';

const geoJsonData = ref(null);
const startNode = ref(null);
const endNode = ref(null);
const shortestPath = ref(null);
const pathWeight = ref(null);
const queryTimeMs = ref(null);
const selectedAlgorithm = ref('ch'); // Default to CH
const verticesMap = ref(new Map());

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
      pathWeight.value = null;
      queryTimeMs.value = null;
    }
  }
};


async function calculateShortestPath() {
  if (!startNode.value || !endNode.value) {
    shortestPath.value = null;
    pathWeight.value = null;
    queryTimeMs.value = null;
    return;
  }
  try {
    const apiUrl = `/api/${selectedAlgorithm.value}/query?from=${startNode.value}&to=${endNode.value}`;
    const response = await fetch(apiUrl);
    if (!response.ok) throw new Error('Failed to fetch shortest path');
    const result = await response.json();
    if (result.path && result.path.length > 0) {
      const pathCoordinates = result.path.map(id => verticesMap.value.get(id)).filter(Boolean);
      if (pathCoordinates.length > 1) {
        shortestPath.value = {
          type: 'Feature',
          geometry: { type: 'LineString', coordinates: pathCoordinates },
          properties: {}
        };
        pathWeight.value = result.weight;
        queryTimeMs.value = result.queryTimeMs;
      }
    } else {
      shortestPath.value = null;
      pathWeight.value = null;
      queryTimeMs.value = null;
    }
  } catch (error) {
    console.error(error);
    shortestPath.value = null;
    pathWeight.value = null;
    queryTimeMs.value = null;
  }
}

watch(endNode, (newValue) => {
  if (newValue) {
    calculateShortestPath();
  }
});

watch(selectedAlgorithm, () => {
  startNode.value = null;
  endNode.value = null;
  shortestPath.value = null;
  pathWeight.value = null;
  queryTimeMs.value = null;
});

onMounted(async () => {
  try {
    const response = await fetch('/api/ch');
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    const graphData = await response.json();
    const features = [];
    const processedVertices = new Set();

    const processGraph = (graph, graphType) => {
      if (!graph) return;
      if (graph.Vertices) {
        for (const vertexId in graph.Vertices) {
          const vertex = graph.Vertices[vertexId];
          if (!verticesMap.value.has(vertex.Id)) {
            verticesMap.value.set(vertex.Id, [vertex.Lon, vertex.Lat]);
          }
          if (!processedVertices.has(vertexId)) {
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
            features.push({
              type: 'Feature',
              geometry: { type: 'LineString', coordinates: [[sourceVertex.Lon, sourceVertex.Lat], [targetVertex.Lon, targetVertex.Lat]] },
              properties: { source: sourceId, target: targetId, type: 'edge', graph: graphType }
            });
          }
        }
      }
    };

    processGraph(graphData.UpwardsGraph, 'upwards');
    processGraph(graphData.DownwardsGraph, 'downwards');

    geoJsonData.value = { type: 'FeatureCollection', features: features };
  } catch (error) {
    console.error('Failed to fetch and parse graph data:', error);
  }
});
</script>

<template>
  <GraphVisualization
    :geo-json-data="geoJsonData"
    :shortest-path="shortestPath"
    :start-node="startNode"
    :end-node="endNode"
    :vertices-map="verticesMap"
    :selected-algorithm="selectedAlgorithm"
    :view-state="viewState"
    @update:view-state="handleViewStateChange"
    @layer-click="handleLayerClick"
  />
  <RouteSelector
    :geo-json-data="geoJsonData"
    :start-node="startNode"
    :end-node="endNode"
    @update:start-node="startNode = $event"
    @update:end-node="endNode = $event"
  />
  <div class="info-overlay">
    <div class="algorithm-selector">
      <label for="algorithm-select">Algorithm:</label>
      <select id="algorithm-select" v-model="selectedAlgorithm">
        <option value="ch">Contraction Hierarchies (CH)</option>
        <option value="cch">Customizable Contraction Hierarchies (CCH)</option>
      </select>
    </div>
    <p v-if="startNode">Start Node: {{ startNode }}</p>
    <p v-if="endNode">End Node: {{ endNode }}</p>
    <p v-if="pathWeight !== null">Path Weight: {{ pathWeight.toFixed(2) }}</p>
    <p v-if="queryTimeMs">Query Time: {{ queryTimeMs.toFixed(2) }} ms</p>
    <p v-if="startNode && !endNode">Click a vertex to select end node.</p>
    <p v-if="!startNode">Click a vertex to select start node.</p>
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
