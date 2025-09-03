
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/PaulMue0/efficient-routeplanning/experiments"
)

func main() {
	experiment := flag.String("experiment", "ch", "The experiment to run (ch, query, cch_preprocess, cch_customization or cch_query)")
	flag.Parse()

	switch *experiment {
	case "ch":
		experiments.RunCHExperiment()
	case "query":
		experiments.RunQueryExperiment()
	case "cch_preprocess":
		experiments.RunCCHPreprocessExperiment()
	case "cch_customization":
		experiments.RunCCHCustomizationExperiment()
	case "cch_query":
		experiments.RunCCHQueryExperiment()
	default:
		fmt.Println("Invalid experiment specified. Use 'ch', 'query', 'cch_preprocess', 'cch_customization' or 'cch_query'.")
		os.Exit(1)
	}
}
