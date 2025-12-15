package huffman

import (
	"bufio"
	"encoding/binary"
	"errors"
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

func _readTree(reader *bitio.Reader, length uint16) (*node, error) {
	var readed uint16 = 0
	var next func() (*node, error)
	next = func() (*node, error) {
		if readed >= length {
			return nil, nil
		}
		bit, err := reader.ReadBit()
		if err != nil {
			return nil, err
		}
		readed += 1
		if bit == 1 {
			if readed+8 > length {
				return nil, errors.New("invalid header structure")
			}
			b, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}
			readed += 8
			return &node{b, 0, nil, nil}, nil
		}
		left, err := next()
		if err != nil {
			return nil, err
		}
		if readed > length {
			return &node{0, 0, left, nil}, nil
		}

		right, err := next()
		if err != nil {
			return nil, err
		}
		return &node{0, 0, left, right}, nil
	}
	return next()
}

func readTree(reader *bitio.Reader) (*node, error) {
	buffer := make([]byte, 2)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		return nil, err
	}
	headerSize := binary.LittleEndian.Uint16(buffer)
	return _readTree(reader, headerSize)
}

func (decoder *HuffmanDecoder) Decode() error {
	reader := bitio.NewReader(decoder.reader)
	root, err := readTree(reader)
	if err != nil {
		if err == io.EOF {
			return errors.New("invalid file structure")
		}
		return err
	}
	if err := reader.Align(); err != nil {
		return err
	}

	buffer := make([]byte, 8)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		return errors.New("invalid file structure")
	}
	length := binary.LittleEndian.Uint64(buffer)
	codes := buildReverseCodes(root)
	writer := bufio.NewWriter(decoder.writer)
	code := ""
	var total uint64 = 0
	for total < length {
		bit, err := reader.ReadBit()
		total += 1
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
