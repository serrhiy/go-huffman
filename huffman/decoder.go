package huffman

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/serrhiy/go-huffman/bitio"
)

type HuffmanDecoder struct {
	reader io.Reader
	writer io.Writer
}

func NewDecoder(reader io.Reader, writer io.Writer) *HuffmanDecoder {
	return &HuffmanDecoder{reader, writer}
}

func _readTree(reader *bitio.Reader) (*node, error) {
	bit, err := reader.ReadBit()
	if err != nil {
		return nil, err
	}
	if bit == 1 {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		return &node{b, 0, nil, nil}, nil
	}
	left, err := _readTree(reader)
	if err != nil {
		if err == io.EOF {
			return &node{0, 0, left, nil}, nil
		}
		return nil, err
	}
	right, err := _readTree(reader)
	if err != nil {
		if err == io.EOF {
			return &node{0, 0, left, right}, nil
		}
		return nil, err
	}
	return &node{0, 0, left, right}, nil
}

func (decoder *HuffmanDecoder) readTree() (*node, error) {
	reader := bufio.NewReader(decoder.reader)
	header, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(header) == 0 {
		return nil, errors.New("input file is empty")
	}
	header = header[0 : len(header)-1]
	root, err := _readTree(bitio.NewReader(bytes.NewReader(header)))
	if err != nil && err != io.EOF {
		return nil, err
	}
	return root, nil
}

func (decoder *HuffmanDecoder) Decode() error {
	root, err := decoder.readTree()
	if err != nil {
		return err
	}
	codes := buildCodes(root)
	for key, value := range codes {
		fmt.Printf("%q -> %s\n", key, value)
	}
	return nil
}
