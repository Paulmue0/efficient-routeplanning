#!/bin/bash

set -e

# Find all files ending in .metis and store them in an array
files=(*.metis)
total_files=${#files[@]}
current_file_num=0

echo "Found $total_files .metis files to process."
echo "=================================================="
total_start_time=$(date +%s)

# Create a main temporary directory for all processing and set up cleanup
MAIN_TEMP_DIR=$(mktemp -d)
echo "Temporary directory created at: $MAIN_TEMP_DIR" # Add this line
trap 'rm -rf "$MAIN_TEMP_DIR"' EXIT

# Use a for loop to iterate over the files
for file in "${files[@]}"; do
  # Check if the glob pattern found any files
  if [ -f "$file" ]; then
    current_file_num=$((current_file_num + 1))
    # Construct the output filename by replacing .metis with .ordering
    output_filename="${file%.metis}.ordering"

    echo "[$current_file_num/$total_files] Processing $file..."
    start_time=$(date +%s)

    # Create a temporary directory for this specific metis file's processing
    FILE_TEMP_DIR=$(mktemp -d "${MAIN_TEMP_DIR}/XXXX")

    # Copy the original metis file into its temporary processing directory
    cp "$file" "${FILE_TEMP_DIR}/input.metis"

    # If an initial .nodes file exists for the original metis file, copy it too
    if [ -f "${file}.nodes" ]; then
      cp "${file}.nodes" "${FILE_TEMP_DIR}/input.metis.nodes"
    fi

    # Run nested_dissection.sh and pipe its output to awk for formatting
    # Pass the path to the copied input file and its temporary directory
    ./nested_dissection.sh "${FILE_TEMP_DIR}/input.metis" "$FILE_TEMP_DIR" | awk '
    {
        lines[NR] = $0
    }
    END {
        print NR # Total number of lines
        for (i = NR; i >= 1; i--) {
            print (NR - i + 1), lines[i] # Rank and node ID
        }
    }' >"$output_filename"

    end_time=$(date +%s)
    duration=$((end_time - start_time))

    echo "[$current_file_num/$total_files] Completed $file in ${duration}s. Output saved to $output_filename"
    echo "--------------------------------------------------"
  fi
done

total_end_time=$(date +%s)
total_duration=$((total_end_time - total_start_time))
echo "All $total_files .metis files have been processed in ${total_duration}s."
