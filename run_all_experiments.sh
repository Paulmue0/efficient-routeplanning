#!/bin/bash

# This script runs all the experiments for the route planning project.

echo "--- Running CH Preprocessing Experiment ---"
go run cmd/ch_experiment/main.go --experiment ch

echo "\n--- Running CH Query Experiment ---"
go run cmd/ch_experiment/main.go --experiment query

echo "\n--- Running CCH Preprocessing Experiment ---"
go run cmd/ch_experiment/main.go --experiment cch_preprocess

echo "\n--- Running CCH Customization Experiment ---"
go run cmd/ch_experiment/main.go --experiment cch_customization

echo "\n--- Running CCH Query Experiment ---"
go run cmd/ch_experiment/main.go --experiment cch_query

echo "\n--- All experiments completed ---"
