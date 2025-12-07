package bitio

import (
	"bufio"
	"io"
)

type Writer struct {
	out bufio.Writer

	cache     byte
	cacheSize byte
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{*bufio.NewWriter(writer), 0, 0}
}

func (writer *Writer) Write(buffer []byte) (int, error) {
	if writer.cacheSize == 0 {
		return writer.out.Write(buffer)
	}
	for index, b := range buffer {
		if err := writer.WriteByte(b); err != nil {
			return index, err
		}
	}
	return len(buffer), nil
}

func (writer *Writer) WriteByte(b byte) error {
	if writer.cacheSize == 0 {
		return writer.out.WriteByte(b)
	}
	err := writer.out.WriteByte(writer.cache | b>>writer.cacheSize)
	if err != nil {
		return err
	}
	writer.cache = b << (8 - writer.cacheSize)
	return nil
}

func (writer *Writer) WriteBits(bits byte, n byte) (err error) {
	bits &= ((1 << n) - 1) << (8 - n)

	if writer.cacheSize+n < 8 {
		writer.cache |= bits << writer.cacheSize
		writer.cacheSize += n
		return
	}

	toWrite := writer.cache | (bits >> writer.cacheSize)
	err = writer.out.WriteByte(toWrite)
	if err != nil {
		return
	}
	writer.cache = bits << (8 - writer.cacheSize)
	writer.cacheSize = writer.cacheSize + n - 8
	return
}

func (writer *Writer) WriteBit(bit bool) error {
	if bit {
		return writer.WriteBits(0b10000000, 1)
	}
	return writer.WriteBits(0, 1)
}

func (writer *Writer) Align() error {
	if writer.cacheSize == 0 {
		return nil
	}
	return writer.WriteBits(0, 8-writer.cacheSize)
}

func (writer *Writer) Flush() error {
	return writer.out.Flush()
}
