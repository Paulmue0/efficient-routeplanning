export async function updateEdgeWeights(updates) {
  try {
    const response = await fetch('/api/cch/update', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    console.error('Error updating edge weights:', error);
    throw error;
  }
}
