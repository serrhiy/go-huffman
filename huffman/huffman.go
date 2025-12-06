package huffman

import (
	"bufio"
	"io"
	"os"
)

const bufferSize = 32 * 1024

type Huffman struct {
	probability map[byte]uint64
}

func NewFromFile(filepath string) (*Huffman, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := make(map[byte]uint64)
	reader := bufio.NewReader(file)
	buffer := make([]byte, bufferSize)
	for {
		bytes, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		for i := range bytes {
			result[buffer[i]]++
		}
	}
	return &Huffman{result}, nil
}
