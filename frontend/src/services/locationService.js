import { geocodeAddress } from './geocodingService';

export async function findClosestVertexToAddress(address, verticesMap) {
  const coords = await geocodeAddress(address);
  if (coords && verticesMap) {
    let closestVertex = null;
    let minDistance = Infinity;
    
    // Iterate through all vertices in the map to find the closest one
    for (const [vertexId, [lon, lat]] of verticesMap) {
      const distance = Math.sqrt(Math.pow(coords.lon - lon, 2) + Math.pow(coords.lat - lat, 2));
      if (distance < minDistance) {
        minDistance = distance;
        closestVertex = {
          properties: {
            id: parseInt(vertexId)
          }
        };
      }
    }
    
    return closestVertex;
  }
  return null; // No coordinates found or no closest vertex
}
