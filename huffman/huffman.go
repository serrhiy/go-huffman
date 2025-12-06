package huffman

import (
	"bufio"
	"fmt"
	"io"
)

const bufferSize = 32 * 1024

type HuffmanEncoder struct {
	reader io.ReadSeeker
	writer io.ReadSeeker
}

func NewEncoder(reader io.ReadSeeker, writer io.ReadSeeker) HuffmanEncoder {
	return HuffmanEncoder{reader, writer}
}

func (encoder HuffmanEncoder) getFrequencyMap() (map[byte]uint, error) {
	encoder.reader.Seek(0, io.SeekStart)

	result := make(map[byte]uint, 1<<8)
	reader := bufio.NewReader(encoder.reader)
	buffer := make([]byte, bufferSize)
	for {
		readed, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		for index := range readed {
			result[buffer[index]] += 1
		}
	}
	return result, nil
}

func (encoder HuffmanEncoder) Encode() error {
	frequencies, err := encoder.getFrequencyMap()
	if err != nil {
		return err
	}
	root := buildTree(frequencies)
	codes := buildCodes(root)

	for char, code := range codes {
		fmt.Printf("%q -> %s\n", char, code)
	}

	return nil
}
