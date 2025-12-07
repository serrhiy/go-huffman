package bitio

import (
	"bufio"
	"fmt"
	"io"
)

type Reader struct {
	in bufio.Reader
	
	cache     byte
	cacheSize byte
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{*bufio.NewReader(reader), 0, 0}
}

func (reader *Reader) ReadBit() (byte, error) {
	if reader.cacheSize > 0 {
		value := (reader.cache & 0b10000000) >> 7
		reader.cache <<= 1
		reader.cacheSize -= 1
		return value, nil
	}
	readed, err := reader.in.ReadByte()
	fmt.Println(readed)
	if err != nil {
		return 0, err
	}
	value := (readed & 0b10000000) >> 7
	reader.cache = readed << 1
	reader.cacheSize = 7
	return value, nil
}

func (reader *Reader) ReadByte() (byte, error) {
	readed, err := reader.in.ReadByte()
	if reader.cacheSize == 0 || err != nil {
		return readed, err
	}
	result := reader.cache | (readed >> reader.cacheSize)
	reader.cache = readed << (8 - reader.cacheSize)
	return result, nil
}

func (reader *Reader) Read(buffer []byte) (int, error) {
	if reader.cacheSize == 0 {
		return reader.in.Read(buffer)
	}
	for index := range buffer {
		if b, err := reader.ReadByte(); err != nil {
			return index, err
		} else {
			buffer[index] = b
		}
	}
	return 0, nil
}
