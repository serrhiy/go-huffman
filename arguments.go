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
		dir := filepath.Dir(input)
		otuputPath := filepath.Join(dir, name+OutputExtension)
		return &arguments{input, otuputPath}, nil
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

func getArguments(encode, decode, output string) (*arguments, error) {
	if len(encode) > 0 && len(decode) > 0 {
		return nil, errors.New("both encode and decode paths are set, specify only one")
	}
	if len(encode) == 0 && len(decode) == 0 {
		return nil, errors.New("either encode or decode path must be specified")
	}
	if len(encode) > 0 {
		return getArgumentsEncode(encode, output)
	}
	return getArgumentsDecode(decode, output)
}
