package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/serrhiy/go-huffman/huffman"
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
		return *input, name + OutputExtension, nil
	}

	return *input, *output, nil
}

func start() error {
	inputFile, outputFile, err := getArguments()
	if err != nil {
		return err
	}

	infile, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer infile.Close()

	outfile, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil
	}
	defer outfile.Close()

	encoder := huffman.NewEncoder(infile, outfile)
	err = encoder.Encode()
	if err != nil {
		return err
	}
	return err
}

func main() {
	flag.Parse()

	err := start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred: %v\n", err)
		os.Exit(1)
	}
}
