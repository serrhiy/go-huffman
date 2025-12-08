package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/serrhiy/go-huffman/huffman"
)

const OutputExtension = ".hfm"

var input = flag.String("input", "", "path to the input file")
var output = flag.String("output", "", "path to the output file")
var service = flag.String("service", "e", "e - encode, d - decode")

func start() error {
	arguments, err := getArguments(*input, *output, *service)
	if err != nil {
		return err
	}

	infile, err := os.Open(arguments.inputFile)
	if err != nil {
		return err
	}
	defer infile.Close()

	outfile, err := os.OpenFile(arguments.outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil
	}
	defer outfile.Close()

	if *service == "e" {
		encoder := huffman.NewEncoder(infile, outfile)
		err = encoder.Encode()
		if err != nil {
			return err
		}
	} else {
		decoder := huffman.NewDecoder(infile, outfile)
		err := decoder.Decode()
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Parse()

	err := start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred: %v\n", err)
		os.Exit(1)
	}
}
