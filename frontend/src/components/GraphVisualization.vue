<script setup>
import { ref, watch } from 'vue';
import { DeckGL, Map as MapComponent } from '@vue-deckgl-suite/maplibre';
import { GeoJsonLayer } from '@vue-deckgl-suite/layers';

const props = defineProps({
  geoJsonData: Object,
  shortestPath: Object,
  startNode: Number,
  endNode: Number,
  verticesMap: Map,
  selectedAlgorithm: String,
  viewState: Object
});

const emit = defineEmits(['update:viewState', 'layerClick']);

const style = 'https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json';

const onViewStateChange = ({ viewState: newViewState }) => {
  emit('update:viewState', newViewState);
};

const getTooltip = ({ object }) => {
  if (!object || !object.properties) return null;
  const { type, id, source, target } = object.properties;
  if (type === 'vertex') return `Vertex: ${id}`;
  if (type === 'edge') return `Edge: ${source} -> ${target}`;
  return null;
};

const handleLayerClick = (info) => {
  emit('layerClick', info);
};

const getFillColor = (feature) => {
  const vertexId = feature.properties.id;
  if (vertexId === props.startNode) return [0, 255, 0, 255]; // Green
  if (vertexId === props.endNode) return [255, 0, 255, 255]; // Magenta
  return [255, 255, 255, 200]; // White
};

const getPointRadius = (feature) => {
  const vertexId = feature.properties.id;
  if (vertexId === props.startNode || vertexId === props.endNode) {
    return 20; // Larger radius for selected nodes
  }
  return 5; // Default radius
};

const getLineColor = (feature) => {
  if (feature.properties.type === 'edge') {
    return feature.properties.graph === 'upwards' ? [255, 0, 0, 255] : [0, 0, 255, 255];
  }
  if (feature.properties.type === 'vertex') {
    const vertexId = feature.properties.id;
    if (vertexId === props.startNode || vertexId === props.endNode) {
      return [0, 0, 0, 255]; // Black stroke for selected vertices
    }
  }
  return [0, 0, 0, 0]; // Transparent for other cases (non-selected vertices, or if no stroke desired)
};

const getLineWidth = (feature) => {
  if (feature.properties.type === 'edge') {
    return 2; // Default width for edges
  }
  if (feature.properties.type === 'vertex') {
    const vertexId = feature.properties.id;
    if (vertexId === props.startNode || vertexId === props.endNode) {
      return 4; // Thicker stroke for selected vertices
    }
  }
  return 0; // No stroke for non-selected vertices (or default to 0 if not explicitly handled)
};
</script>

<template>
  <DeckGL :get-tooltip="getTooltip" :view-state="props.viewState" @view-state-change="onViewStateChange">
    <MapComponent height="100vh" :style :center="[props.viewState.longitude, props.viewState.latitude]" :zoom="props.viewState.zoom" />
    <GeoJsonLayer v-if="props.geoJsonData" id="graph-layer" :data="props.geoJsonData" pointType="circle" :filled="true"
      :stroked="true" :pickable="true" :getFillColor="getFillColor" :getLineColor="getLineColor"
      @click="handleLayerClick" :getLineWidth="getLineWidth" lineWidthUnits="pixels" :getPointRadius="getPointRadius"
      pointRadiusUnits="pixels"
      :update-triggers="{ getFillColor: [props.startNode, props.endNode], getPointRadius: [props.startNode, props.endNode], getLineColor: [props.startNode, props.endNode], getLineWidth: [props.startNode, props.endNode] }" />
    <GeoJsonLayer v-if="props.shortestPath" id="shortest-path-layer" :data="props.shortestPath" :getLineColor="[0, 255, 0, 255]"
      :getLineWidth="5" lineWidthUnits="pixels" />
  </DeckGL>
</template>

<style lang="scss">
@import 'maplibre-gl/dist/maplibre-gl.css';
</style>
