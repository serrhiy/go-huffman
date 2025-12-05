package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const OutputExtension = ".hfm"

var input = flag.String("input", "", "path to the input file")
var output = flag.String("output", "", "path to the output file")

func getArguments() (string, string, error) {
	if *input == "" {
		return "", "", errors.New("input argument is mandatory")
	}

	if *output == "" {
		base := filepath.Base(*input)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		return *input, fmt.Sprintf("%s%s", name, OutputExtension), nil
	}

	return *input, *output, nil
}

func main() {
	flag.Parse()

	inputFile, outputFile, err := getArguments()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Passed args are %s, %s\n", inputFile, outputFile)
}
