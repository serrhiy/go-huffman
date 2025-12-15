package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/serrhiy/go-huffman/huffman"
)

const OutputExtension = ".hfm"

var output = flag.String("o", "", "path to the output file")
var encode = flag.String("e", "", "encode file")
var decode = flag.String("d", "", "decode file")

func start() error {
	arguments, err := getArguments(*encode, *decode, *output)
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

	if len(*encode) > 0 {
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

	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred: %v\n", err)
		if err == huffman.ErrInvalidStructure {
			os.Remove(*output)
		}
		os.Exit(1)
	}
}
