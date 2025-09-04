<script setup>
import { ref, watch } from 'vue';
import { findClosestVertexToAddress } from '../services/locationService';

const props = defineProps({
  geoJsonData: Object,
  verticesMap: Map,
  startNode: Number,
  endNode: Number,
});

const emit = defineEmits(['update:startNode', 'update:endNode']);

const startAddress = ref('');
const endAddress = ref('');

const startVertexId = ref(null);
const endVertexId = ref(null);

watch(() => props.startNode, (newVal) => {
  startVertexId.value = newVal;
});

watch(() => props.endNode, (newVal) => {
  endVertexId.value = newVal;
});

async function handleStartAddressInput() {
  if (startAddress.value.trim() === '') {
    startVertexId.value = null;
    emit('update:startNode', null);
    return;
  }
  
  // Check if input is a number (vertex ID)
  const numericId = parseInt(startAddress.value.trim());
  if (!isNaN(numericId) && props.verticesMap && props.verticesMap.has(numericId)) {
    startVertexId.value = numericId;
    emit('update:startNode', numericId);
    return;
  }
  
  // Otherwise, treat as address and geocode
  const closestVertex = await findClosestVertexToAddress(startAddress.value, props.verticesMap);
  if (closestVertex) {
    startVertexId.value = closestVertex.properties.id;
    emit('update:startNode', closestVertex.properties.id);
  } else {
    startVertexId.value = null;
    emit('update:startNode', null);
    console.warn('Could not find a vertex for start address:', startAddress.value);
  }
}

async function handleEndAddressInput() {
  if (endAddress.value.trim() === '') {
    endVertexId.value = null;
    emit('update:endNode', null);
    return;
  }
  
  // Check if input is a number (vertex ID)
  const numericId = parseInt(endAddress.value.trim());
  if (!isNaN(numericId) && props.verticesMap && props.verticesMap.has(numericId)) {
    endVertexId.value = numericId;
    emit('update:endNode', numericId);
    return;
  }
  
  // Otherwise, treat as address and geocode
  const closestVertex = await findClosestVertexToAddress(endAddress.value, props.verticesMap);
  if (closestVertex) {
    endVertexId.value = closestVertex.properties.id;
    emit('update:endNode', closestVertex.properties.id);
  } else {
    endVertexId.value = null;
    emit('update:endNode', null);
    console.warn('Could not find a vertex for end address:', endAddress.value);
  }
}
</script>

<template>
  <div class="route-selector">
    <div class="input-group">
      <label for="start-address">Start:</label>
      <input type="text" id="start-address" v-model="startAddress" @keyup.enter="handleStartAddressInput" placeholder="Enter start address or vertex ID">
      <span v-if="startVertexId !== null"> (ID: {{ startVertexId }})</span>
    </div>
    <div class="input-group">
      <label for="end-address">Destination:</label>
      <input type="text" id="end-address" v-model="endAddress" @keyup.enter="handleEndAddressInput" placeholder="Enter destination address or vertex ID">
      <span v-if="endVertexId !== null"> (ID: {{ endVertexId }})</span>
    </div>
  </div>
</template>

<style scoped>
.route-selector {
  background: rgba(255, 255, 255, 0.8);
  padding: 10px;
  border-radius: 5px;
  z-index: 1000;
  color: black;
  font-family: sans-serif;
  position: absolute;
  top: 10px;
  right: 10px;
}

.input-group {
  margin-bottom: 10px;
}

.input-group label {
  display: block;
  margin-bottom: 5px;
}

.input-group input[type="text"] {
  width: 200px;
  padding: 5px;
  border: 1px solid #ccc;
  border-radius: 3px;
}
</style>
