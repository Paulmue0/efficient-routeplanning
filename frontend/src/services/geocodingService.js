export async function geocodeAddress(address) {
  const url = `https://nominatim.openstreetmap.org/search?q=${encodeURIComponent(address)}&format=jsonv2&polygon_geojson=1`;
  try {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    if (data && data.length > 0) {
      // Return lon and lat of the first result
      return { lon: parseFloat(data[0].lon), lat: parseFloat(data[0].lat) };
    } else {
      return null; // No results found
    }
  } catch (error) {
    console.error('Error during geocoding:', error);
    return null;
  }
}
