package huffman

import (
	"bufio"
	"bytes"
	"io"

	"github.com/serrhiy/go-huffman/bitio"
)

type HuffmanDecoder struct {
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewDecoder(reader io.Reader, writer io.Writer) *HuffmanDecoder {
	return &HuffmanDecoder{bufio.NewReader(reader), bufio.NewWriter(writer)}
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
			return nil, nil
		}
		return nil, err
	}
	return &node{0, 0, left, right}, nil
}

func (decoder *HuffmanDecoder) readTree() (*node, error) {
	header, err := decoder.reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	if len(header) == 0 {
		return nil, nil
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
	if err != nil{
		return err
	}
	if root == nil {
		return nil
	}
	codes := buildReverseCodes(root)
	reader := bitio.NewReader(decoder.reader)
	writer := bufio.NewWriter(decoder.writer)
	code := ""
	for {
		bit, err := reader.ReadBit()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if bit == 1 {
			code += "1"
		} else {
			code += "0"
		}
		if char, ok := codes[code]; ok {
			writer.WriteByte(char)
			code = ""
		}
	}
	return writer.Flush()
}
