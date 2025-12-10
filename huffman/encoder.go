package huffman

import (
	"bufio"
	"io"

	"github.com/serrhiy/go-huffman/bitio"
)

const bufferSize = 32 * 1024

type HuffmanEncoder struct {
	reader io.ReadSeeker
	writer io.Writer
}

func NewEncoder(reader io.ReadSeeker, writer io.Writer) *HuffmanEncoder {
	return &HuffmanEncoder{reader, writer}
}

func (encoder *HuffmanEncoder) getFrequencyMap() (map[byte]uint, error) {
	if _, err := encoder.reader.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

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
	if len(frequencies) == 0 {
		return nil
	}
	root := buildTree(frequencies)
	codes := buildCodes(root)
	if err := encoder.writeCodes(root); err != nil {
		return err
	}
	if err := encoder.encodeContent(codes); err != nil {
		return err
	}
	return nil
}

func (encoder *HuffmanEncoder) writeCodes(root *node) error {
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

func (encoder *HuffmanEncoder) encodeContent(codes map[byte]string) error {
	if _, err := encoder.reader.Seek(0, io.SeekStart); err != nil {
		return err
	}
	reader := bufio.NewReader(encoder.reader)
	writer := bitio.NewWriter(encoder.writer)
	buffer := make([]byte, bufferSize)
	for {
		readed, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		for i := range readed {
			code := codes[buffer[i]]
			for j := range code {
				if code[j] == '1' {
					err = writer.WriteBit(1)
				} else {
					err = writer.WriteBit(0)
				}
				if err != nil {
					return err
				}
			}
		}
	}
	return writer.Flush()
}
