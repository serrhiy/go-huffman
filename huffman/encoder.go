package huffman

import (
	"bufio"
	"encoding/binary"
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
	length, _ := calculateContentSize(codes, frequencies)
	if err := encoder.writeCodes(root); err != nil {
		return err
	}
	if err := encoder.encodeContent(codes, length); err != nil {
		return err
	}
	return nil
}

func (encoder *HuffmanEncoder) writeCodes(root *node) error {
	b := make([]byte, 2)
	treeSize := calculateTreeSize(root)
	binary.LittleEndian.PutUint16(b, treeSize)

	bitWriter := bitio.NewWriter(encoder.writer)
	if _, err := bitWriter.Write(b); err != nil {
		return err
	}
	if err := writeCodes(root, bitWriter); err != nil {
		return err
	}
	if err := bitWriter.Flush(); err != nil {
		return err
	}
	return nil
}

func (encoder *HuffmanEncoder) encodeContent(codes map[byte]string, length uint64) error {
	if _, err := encoder.reader.Seek(0, io.SeekStart); err != nil {
		return err
	}
	reader := bufio.NewReader(encoder.reader)
	writer := bitio.NewWriter(encoder.writer)
	buffer := make([]byte, bufferSize)

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, length)
	if _, err := writer.Write(b); err != nil {
		return err
	}

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
