package main

import (
	"errors"
	"path/filepath"
)

type arguments struct {
	inputFile  string
	outputFile string
}

func getArgumentsEncode(input, output string) (*arguments, error) {
	if input == "" {
		return nil, errors.New("input argument is mandatory")
	}
	if output == "" {
		base := filepath.Base(input)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		return &arguments{input, name + OutputExtension}, nil
	}
	return &arguments{input, output}, nil
}

func getArgumentsDecode(input, output string) (*arguments, error) {
	if input == "" {
		return nil, errors.New("input argument is mandatory")
	}
	if output == "" {
		return nil, errors.New("output argument is mandatory")
	}
	return &arguments{input, output}, nil
}

var parsers = map[string]func(string, string) (*arguments, error){
	"e": getArgumentsEncode,
	"d": getArgumentsDecode,
}

func getArguments(input, output, service string) (*arguments, error) {
	parser, ok := parsers[service]
	if !ok {
		return nil, errors.New("invalid service parameter: " + service)
	}
	return parser(input, output)
}
