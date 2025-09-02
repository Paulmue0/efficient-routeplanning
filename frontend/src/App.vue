<script setup>
import { ref, onMounted } from 'vue';
import { DeckGL, Map } from '@vue-deckgl-suite/maplibre';
import { GeoJsonLayer } from '@vue-deckgl-suite/layers';

const style = 'https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json';
const geoJsonData = ref(null);

const getTooltip = ({ object }) => {
  if (!object || !object.properties) {
    return null;
  }
  const { type, id, source, target } = object.properties;
  if (type === 'vertex') {
    return `Vertex: ${id}`;
  }
  if (type === 'edge') {
    return `Edge: ${source} -> ${target}`;
  }
  return null;
};

onMounted(async () => {
  try {
    const response = await fetch('/api/ch'); // Or /api/cch
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const graphData = await response.json();
    const features = [];
    const processedVertices = new Set();

    const processGraph = (graph, graphType) => {
      if (!graph) return;

      // Process vertices
      if (graph.Vertices) {
        for (const vertexId in graph.Vertices) {
          if (!processedVertices.has(vertexId)) {
            const vertex = graph.Vertices[vertexId];
            features.push({
              type: 'Feature',
              geometry: {
                type: 'Point',
                coordinates: [vertex.Lon, vertex.Lat]
              },
              properties: {
                id: vertex.Id,
                type: 'vertex'
              }
            });
            processedVertices.add(vertexId);
          }
        }
      }

      // Process edges
      if (graph.Edges && graph.Vertices) {
        for (const sourceId in graph.Edges) {
          const sourceVertex = graph.Vertices[sourceId];
          if (!sourceVertex) continue;

          for (const targetId in graph.Edges[sourceId]) {
            const targetVertex = graph.Vertices[targetId];
            if (!targetVertex) continue;

            features.push({
              type: 'Feature',
              geometry: {
                type: 'LineString',
                coordinates: [
                  [sourceVertex.Lon, sourceVertex.Lat],
                  [targetVertex.Lon, targetVertex.Lat]
                ]
              },
              properties: {
                source: sourceId,
                target: targetId,
                type: 'edge',
                graph: graphType
              }
            });
          }
        }
      }
    };

    processGraph(graphData.UpwardsGraph, 'upwards');
    processGraph(graphData.DownwardsGraph, 'downwards');

    geoJsonData.value = {
      type: 'FeatureCollection',
      features: features
    };
  } catch (error) {
    console.error('Failed to fetch and parse graph data:', error);
  }
});
</script>

<template>
  <DeckGL :get-tooltip="getTooltip">
    <Map height="100vh" :style :center="[9.244557, 48.667421]" :zoom="11" />
    <GeoJsonLayer
      v-if="geoJsonData"
      id="graph-layer"
      :data="geoJsonData"
      pointType="circle"
      :filled="true"
      :stroked="false"
      :pickable="true"
      :getFillColor="[255, 255, 255, 200]"
      :getLineColor="feature => feature.properties.graph === 'upwards' ? [255, 0, 0, 255] : [0, 0, 255, 255]"
      :getLineWidth="2"
      lineWidthUnits="pixels"
      :getPointRadius="2"
      pointRadiusUnits="pixels"
    />
  </DeckGL>
</template>

<style lang="scss">
@import 'maplibre-gl/dist/maplibre-gl.css';
</style>