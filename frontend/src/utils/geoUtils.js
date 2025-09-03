import * as turf from '@turf/turf';

export function findClosestPoints(lon, lat, n, geoJsonData) {
  if (!geoJsonData || !geoJsonData.features || geoJsonData.features.length === 0) {
    return [];
  }

  const searchPoint = turf.point([lon, lat]);
  const distances = [];

  // Filter for vertex features and calculate distances
  for (const feature of geoJsonData.features) {
    if (feature.properties && feature.properties.type === 'vertex') {
      const vertexPoint = turf.point(feature.geometry.coordinates);
      const distance = turf.distance(searchPoint, vertexPoint, { units: 'kilometers' });
      distances.push({ feature, distance });
    }
  }

  // Sort by distance and return the top n
  distances.sort((a, b) => a.distance - b.distance);
  return distances.slice(0, n).map(item => item.feature);
}
