<script setup>
import { ref, watch } from 'vue';
import { DeckGL, Map as MapComponent } from '@vue-deckgl-suite/maplibre';
import { GeoJsonLayer, ArcLayer } from '@vue-deckgl-suite/layers';

const hoverInfo = ref(null);

const props = defineProps({
  geoJsonData: Object,
  shortestPath: Object,
  pathShortcuts: Object, // New prop for path shortcuts
  startNode: Number,
  endNode: Number,
  verticesMap: Map,
  selectedAlgorithm: String,
  viewState: Object,
  blockedEdges: Array,
  shortcutsGeoJsonData: Object, // New prop
  showShortcuts: Boolean, // New prop
  baseGeoJsonData: Object // New prop
});

const emit = defineEmits(['update:viewState', 'layerClick', 'edgeClick']);

const style = 'https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json';

const onViewStateChange = ({ viewState: newViewState }) => {
  emit('update:viewState', newViewState);
};

const getTooltip = ({ object }) => {
  if (!object || !object.properties) return null;
  const { type, id, source, target, weight } = object.properties;
  if (type === 'vertex') return `Vertex: ${id}`;
  if (type === 'edge') return `Edge: ${source} -> ${target}, Weight: ${weight}`;
  return null;
};

const handleLayerClick = (info) => {
  if (info.object && info.object.properties.type === 'vertex') {
    emit('layerClick', info);
  } else if (info.object && info.object.properties.type === 'edge') {
    emit('edgeClick', { from: Number(info.object.properties.source), to: Number(info.object.properties.target) });
  }
};

const getFillColor = (feature) => {
  const vertexId = feature.properties.id;
  if (vertexId === props.startNode) return [0, 255, 0, 255]; // Green
  if (vertexId === props.endNode) return [255, 0, 255, 255]; // Magenta
  if (hoverInfo.value && hoverInfo.value.object && hoverInfo.value.object.properties.type === 'vertex' && hoverInfo.value.object.properties.id === vertexId) {
    return [0, 255, 255, 255]; // Cyan for hovered vertex
  }
  return [255, 255, 255, 200]; // White
};

const getPointRadius = (feature) => {
  const vertexId = feature.properties.id;
  if (vertexId === props.startNode || vertexId === props.endNode) {
    return 20; // Larger radius for selected nodes
  }
  if (hoverInfo.value && hoverInfo.value.object && hoverInfo.value.object.properties.type === 'vertex' && hoverInfo.value.object.properties.id === vertexId) {
    return 10; // Slightly larger for hovered vertex
  }
  return 5; // Default radius
};

const getLineColor = (feature) => {
  if (feature.properties.type === 'edge') {
    // Check if the edge is blocked
    const isBlocked = props.blockedEdges.some(be => be.from === feature.properties.source && be.to === feature.properties.target);
    if (isBlocked) {
      return [255, 100, 0, 255]; // Orange for blocked edges
    }
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
    const isBlocked = props.blockedEdges.some(be => be.from === feature.properties.source && be.to === feature.properties.target);
    if (isBlocked) {
      return 4; // Thicker for blocked edges
    }
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
const getShortestPathLineColor = (feature) => {
  // Since shortcuts are now shown as arcs, shortest path layer only shows real edges
  return [0, 255, 0, 255]; // Green for real path segments
};

const getShortestPathLineWidth = (feature) => {
  // Since shortcuts are now shown as arcs, shortest path layer only shows real edges
  return 5; // Default width for real path segments
};
</script>

<template>
  <DeckGL :get-tooltip="getTooltip" :view-state="props.viewState" @view-state-change="onViewStateChange"
    @on-hover="hoverInfo = $event">
    <MapComponent height="100vh" :style :center="[props.viewState.longitude, props.viewState.latitude]"
      :zoom="props.viewState.zoom" />
    <GeoJsonLayer v-if="props.baseGeoJsonData" id="graph-layer" :data="props.baseGeoJsonData" pointType="circle"
      :filled="true" :stroked="true" :pickable="true" :getFillColor="getFillColor" :getLineColor="getLineColor"
      @click="handleLayerClick" :getLineWidth="getLineWidth" lineWidthUnits="pixels" :getPointRadius="getPointRadius"
      pointRadiusUnits="pixels"
      :update-triggers="{ getFillColor: [props.startNode, props.endNode, hoverInfo], getPointRadius: [props.startNode, props.endNode, hoverInfo], getLineColor: [props.startNode, props.endNode, props.blockedEdges], getLineWidth: [props.startNode, props.endNode, props.blockedEdges] }" />
    <ArcLayer
      v-if="(props.selectedAlgorithm === 'ch' || props.selectedAlgorithm === 'cch') && props.shortcutsGeoJsonData && props.showShortcuts"
      id="shortcuts-arc-layer" :data="props.shortcutsGeoJsonData.features"
      :getSourcePosition="d => d.geometry.coordinates[0]" :getTargetPosition="d => d.geometry.coordinates[1]"
      :getSourceColor="[255, 255, 0, 255]" :getTargetColor="[255, 255, 0, 255]" :getWidth="2" widthUnits="pixels"
      :pickable="false" />
    <GeoJsonLayer v-if="props.shortestPath" id="shortest-path-layer" :data="props.shortestPath"
      :getLineColor="getShortestPathLineColor" :getLineWidth="getShortestPathLineWidth" lineWidthUnits="pixels" />
    <ArcLayer v-if="props.pathShortcuts" id="path-shortcuts-arc-layer" :data="props.pathShortcuts.features"
      :getSourcePosition="d => d.geometry.coordinates[0]" :getTargetPosition="d => d.geometry.coordinates[1]"
      :getSourceColor="[255, 165, 0, 255]" :getTargetColor="[255, 165, 0, 255]" :getWidth="4" widthUnits="pixels"
      :pickable="true" />
  </DeckGL>
</template>

<style lang="scss">
@import 'maplibre-gl/dist/maplibre-gl.css';
</style>
