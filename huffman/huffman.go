package huffman

import (
	"bufio"
	"io"

	"github.com/serrhiy/go-huffman/bitio"
)

const bufferSize = 32 * 1024

type HuffmanEncoder struct {
	reader io.ReadSeeker
	writer io.WriteSeeker
}

func NewEncoder(reader io.ReadSeeker, writer io.WriteSeeker) *HuffmanEncoder {
	return &HuffmanEncoder{reader, writer}
}

func (encoder *HuffmanEncoder) getFrequencyMap() (map[byte]uint, error) {
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

func (encoder *HuffmanEncoder) Encode() error {
	frequencies, err := encoder.getFrequencyMap()
	if err != nil {
		return err
	}
	root := buildTree(frequencies)
	err = encoder.writeCodes(root)
	if err != nil {
		return err
	}
	return nil
}

func (encoder *HuffmanEncoder) writeCodes(root *node) error {
	_, err := encoder.writer.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	bitWriter := bitio.NewWriter(encoder.writer)
	if err := writeCodes(root, bitWriter); err != nil {
		return err
	}
	if err := bitWriter.Align(); err != nil {
		return err
	}
	if err := bitWriter.WriteByte('\n'); err != nil {
		return err
	}
	if err := bitWriter.Flush(); err != nil {
		return err
	}
	return nil
}
