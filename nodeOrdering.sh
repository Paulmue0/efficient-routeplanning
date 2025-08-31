# This script finds all .metis files and runs the KaHIP node_ordering tool on each.

# Use a for loop to find all files ending in .metis
for file in *.metis; do
  # Check if the glob pattern found any files
  if [ -f "$file" ]; then
    # Construct the output filename by replacing .metis with .ordering
    output_filename="${file%.metis}.ordering.txt"

    echo "Processing $file..."

    ./KaHIP/deploy/node_ordering "$file" --output_filename="$output_filename"

    echo "Completed $file. Output saved to $output_filename"
    echo "--------------------------------------------------"
  fi
done

echo "All .metis files have been processed."
