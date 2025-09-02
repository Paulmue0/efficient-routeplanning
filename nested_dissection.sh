#!/bin/bash

set -e

METIS_FILE=$1
CURRENT_TEMP_DIR=$2 # New argument for the current temporary directory
KAHIP_DIR="./KaHIP/deploy"
MIN_PARTITION_SIZE=10

# Ensure the current temporary directory exists
mkdir -p "$CURRENT_TEMP_DIR"

if [ ! -f "$METIS_FILE" ]; then
    echo "File not found: $METIS_FILE" >&2
    exit 1
fi

NUM_NODES=$(head -n 1 "$METIS_FILE" | awk '{print $1}')

# Base case of the recursion
if [ "$NUM_NODES" -le "$MIN_PARTITION_SIZE" ]; then
    if [ -f "${METIS_FILE}.nodes" ]; then
        cat "${METIS_FILE}.nodes"
    else
        seq 1 "$NUM_NODES"
    fi
    exit 0
fi

# All intermediate files are now created in CURRENT_TEMP_DIR
SEPARATOR_FILE="${CURRENT_TEMP_DIR}/separator"

# Run node_separator
"$KAHIP_DIR/node_separator" "$METIS_FILE" --output_filename="$SEPARATOR_FILE" >/dev/null 2>/dev/null

PART_0_NODES="${CURRENT_TEMP_DIR}/p0.nodes"
PART_1_NODES="${CURRENT_TEMP_DIR}/p1.nodes"
SEPARATOR_NODES="${CURRENT_TEMP_DIR}/sep.nodes"

# Touch the files to ensure they exist
touch "$PART_0_NODES" "$PART_1_NODES" "$SEPARATOR_NODES"

# Create files with the node IDs for each partition
awk -v p0="$PART_0_NODES" -v p1="$PART_1_NODES" -v sep="$SEPARATOR_NODES" \
    '{ 
        if ($1 == 0) { print NR >> p0 } 
        else if ($1 == 1) { print NR >> p1 } 
        else { print NR >> sep } 
    }' "$SEPARATOR_FILE"

PART_0_METIS="${CURRENT_TEMP_DIR}/p0.metis"
PART_1_METIS="${CURRENT_TEMP_DIR}/p1.metis"

# Create subgraph METIS files using the Python helper script
if [ -s "$PART_0_NODES" ]; then
    python create_subgraph.py "$METIS_FILE" "$PART_0_NODES" "$PART_0_METIS"
fi
if [ -s "$PART_1_NODES" ]; then
    python create_subgraph.py "$METIS_FILE" "$PART_1_NODES" "$PART_1_METIS"
fi

# Create the .nodes files for the subgraphs to track original node IDs
if [ -f "${METIS_FILE}.nodes" ]; then
    if [ -s "$PART_0_NODES" ]; then
        awk 'FNR==NR{a[$1];next} FNR in a' "$PART_0_NODES" "${METIS_FILE}.nodes" > "${PART_0_METIS}.nodes"
    fi
    if [ -s "$PART_1_NODES" ]; then
        awk 'FNR==NR{a[$1];next} FNR in a' "$PART_1_NODES" "${METIS_FILE}.nodes" > "${PART_1_METIS}.nodes"
    fi
else
    if [ -s "$PART_0_NODES" ]; then
        cp "$PART_0_NODES" "${PART_0_METIS}.nodes"
    fi
    if [ -s "$PART_1_NODES" ]; then
        cp "$PART_1_NODES" "${PART_1_METIS}.nodes"
    fi
fi

# Print separator nodes first
if [ -s "$SEPARATOR_NODES" ]; then
    if [ -f "${METIS_FILE}.nodes" ]; then
        # Map separator nodes to original IDs
        awk 'FNR==NR{a[FNR]=$0;next} {print a[$1]}' "${METIS_FILE}.nodes" "$SEPARATOR_NODES" 
    else
        cat "$SEPARATOR_NODES"
    fi
fi

# Recursive calls
NEXT_TEMP_DIR_0=$(mktemp -d "${CURRENT_TEMP_DIR}/XXXX")
NEXT_TEMP_DIR_1=$(mktemp -d "${CURRENT_TEMP_DIR}/XXXX")

if [ -f "$PART_0_METIS" ]; then
    ./nested_dissection.sh "$PART_0_METIS" "$NEXT_TEMP_DIR_0"
fi
if [ -f "$PART_1_METIS" ]; then
    ./nested_dissection.sh "$PART_1_METIS" "$NEXT_TEMP_DIR_1"
fi

# No cleanup here, the top-level script will handle it.
