// main.ts or main.js
import { createApp } from 'vue';
import App from './App.vue';
import { Map } from '@vue-deckgl-suite/maplibre';
import { ArcLayer } from '@vue-deckgl-suite/layers';

const app = createApp(App);

// Register components
app.component('MaplibreMap', Map);
app.component('ArcLayer', ArcLayer);

app.mount('#app');
