
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/PaulMue0/efficient-routeplanning/experiments"
)

func main() {
	experiment := flag.String("experiment", "ch", "The experiment to run (ch, query or cch_preprocess)")
	flag.Parse()

	switch *experiment {
	case "ch":
		experiments.RunCHExperiment()
	case "query":
		experiments.RunQueryExperiment()
	case "cch_preprocess":
		experiments.RunCCHPreprocessExperiment()
	default:
		fmt.Println("Invalid experiment specified. Use 'ch', 'query' or 'cch_preprocess'.")
		os.Exit(1)
	}
}
