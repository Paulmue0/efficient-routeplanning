import { geocodeAddress } from './geocodingService';
import { findClosestPoints } from '../utils/geoUtils';

export async function findClosestVertexToAddress(address, geoJsonData) {
  const coords = await geocodeAddress(address);
  if (coords) {
    // Find the single closest point (n=1)
    const closestVertices = findClosestPoints(coords.lon, coords.lat, 1, geoJsonData);
    if (closestVertices && closestVertices.length > 0) {
      return closestVertices[0];
    }
  }
  return null; // No coordinates found or no closest vertex
}
