package huffman

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

// func bitsNumberToBytesNumber(bits uint16) uint16 {
// 	return (bits + 8 - bits%8) / 8
// }

type brokenReader struct {
	io.ReadSeeker
}

func (b *brokenReader) Read([]byte) (int, error) {
	return 0, errors.New("read error")
}

func (b *brokenReader) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

type failingWriter struct {
	n int
}

func (fw *failingWriter) Write(p []byte) (int, error) {
	if fw.n <= 0 {
		return 0, errors.New("write failed")
	}
	if len(p) > fw.n {
		p = p[:fw.n]
	}
	fw.n -= len(p)
	return len(p), nil
}

// func TestWriteCodesSingleLeaf(t *testing.T) {
// 	root := &node{
// 		left: &node{
// 			char:  'a',
// 			count: 1,
// 		},
// 		right: nil,
// 	}

// 	buf := &bytes.Buffer{}
// 	enc := NewEncoder(nil, buf)

// 	if err := enc.writeCodes(root); err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	out := buf.Bytes()
// 	expectedLength := calculateTreeSize(root)
// 	expexted := bitsNumberToBytesNumber(expectedLength) + 2
// 	if len(out) != int(expexted) {
// 		t.Fatalf("expected size %d, got %d", expexted, len(out))
// 	}
// 	actualLength := binary.LittleEndian.Uint16(out)
// 	if expectedLength != actualLength {
// 		t.Fatalf("expected header size: %d, got: %d", expectedLength, actualLength)
// 	}
// }

// func TestWriteCodesSimpleTree(t *testing.T) {
// 	root := &node{
// 		left:  &node{char: 'A', count: 3},
// 		right: &node{char: 'B', count: 5},
// 	}

// 	buf := &bytes.Buffer{}
// 	enc := NewEncoder(nil, buf)

// 	err := enc.writeCodes(root)
// 	if err != nil {
// 		t.Fatalf("writeCodes returned error: %v", err)
// 	}

// 	b := buf.Bytes()
// 	if len(b) == 0 {
// 		t.Fatalf("expected non empty result, got %v", b)
// 	}

// 	expectedLength := calculateTreeSize(root)
// 	actualLength := binary.LittleEndian.Uint16(b)

// 	if expectedLength != actualLength {
// 		t.Fatalf("expected some bit-encoded data before newline")
// 	}
// }

// func TestWriteCodesWriterError(t *testing.T) {
// 	root := &node{char: 'X', count: 1}

// 	fw := &failingWriter{n: 0}
// 	enc := NewEncoder(nil, fw)

// 	err := enc.writeCodes(root)
// 	if err == nil {
// 		t.Fatalf("expected writer error")
// 	}
// }

func TestEncodeContentSingleByte(t *testing.T) {
	reader := bytes.NewReader([]byte{'A'})
	writer := &bytes.Buffer{}
	enc := NewEncoder(reader, writer)
	codes := map[byte]string{'a': "1"}
	size, _ := calculateContentSize(codes, map[byte]uint{'a': 1})

	if err := enc.encodeContent(codes, size); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := writer.Bytes()
	if actualSize := binary.LittleEndian.Uint64(data); actualSize != size {
		t.Fatalf("invalid content size, expected %d, got %d", size, actualSize)
	}
}

func TestEncodeContentMultipleBytes(t *testing.T) {
	reader := bytes.NewReader([]byte{'a', 'b', 'a'})
	writer := &bytes.Buffer{}
	enc := NewEncoder(reader, writer)

	codes := map[byte]string{'a': "0", 'b': "1"}
	frequencies := map[byte]uint{'A': 2, 'B': 1}
	size, _ := calculateContentSize(codes, frequencies)

	if err := enc.encodeContent(codes, size); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := writer.Bytes()
	if len(data) == 0 {
		t.Fatal("expected some bytes to be written")
	}
}

func TestEncodeContentVariedBitLengths(t *testing.T) {
	reader := bytes.NewReader([]byte{'a', 'b', 'c'})
	writer := &bytes.Buffer{}
	enc := NewEncoder(reader, writer)

	codes := map[byte]string{'A': "1", 'B': "01", 'C': "001"}

	if err := enc.encodeContent(codes, 1+2+3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if writer.Len() == 0 {
		t.Fatal("expected data to be written")
	}
}

func TestEncodeContentWriterError(t *testing.T) {
	reader := bytes.NewReader([]byte{'A'})
	enc := NewEncoder(reader, &failingWriter{})

	codes := map[byte]string{'A': "1"}
	if err := enc.encodeContent(codes, 1); err == nil {
		t.Fatal("expected write error, got nil")
	}
}

func TestEncodeDeterministicLength(t *testing.T) {
	input := []byte("this is a test string for huffman encoder")
	repetitions := 5
	var lastLen int

	for i := range repetitions {
		r := bytes.NewReader(input)
		w := &bytes.Buffer{}
		enc := NewEncoder(r, w)

		if err := enc.Encode(); err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		encodedLen := w.Len()
		if i > 0 && encodedLen != lastLen {
			t.Fatalf("encoded length differs on repetition %d: got %d, want %d", i, encodedLen, lastLen)
		}

		lastLen = encodedLen
	}
}
