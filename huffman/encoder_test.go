package huffman

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/serrhiy/go-huffman/bitio"
)

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

func TestWriteHeader(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		writer := &bytes.Buffer{}
		reader := bytes.NewReader([]byte{})
		encoder := NewEncoder(reader, writer)
		if err := encoder.writeHeader(nil); err != nil {
			t.Fatalf("unexpected error while writing header: %v", err)
		}
		header := writer.Bytes()
		if len(header) != 2 {
			t.Fatalf("length of empty byffer must be 2, got: %d", len(header))
		}
		if header[0] != 0 || header[1] != 0 {
			t.Fatalf("header length for empty buffer mast be 0, actual: %v", header)
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		root := &node{
			left: &node{char: 'a', count: 10},
		}
		fw := &failingWriter{n: 3}
		enc := NewEncoder(nil, fw)
		err := enc.writeHeader(root)
		if err == nil {
			t.Fatalf("expected writer error")
		}
	})

	t.Run("single leaf", func(t *testing.T) {
		root := &node{
			left: &node{char: 'a', count: 10},
		}
		writer := &bytes.Buffer{}
		reader := bytes.NewReader([]byte{})
		encoder := NewEncoder(reader, writer)
		if err := encoder.writeHeader(root); err != nil {
			t.Fatalf("unexpected error while writing header: %v", err)
		}
		header := writer.Bytes()
		if len(header) != 4 {
			t.Fatalf("invalid header length, expected: %d, got: %d", 4, len(header))
		}
		expectedLength := calculateTreeSize(root)
		length := binary.LittleEndian.Uint16(header)
		if length != expectedLength {
			t.Fatalf("invalid length field, expected: %d, got: %d", expectedLength, length)
		}
		codes := func() []byte {
			buf := &bytes.Buffer{}
			writer := bitio.NewWriter(buf)
			if err := writeCodes(root, writer); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			writer.Flush()
			return buf.Bytes()
		}()

		if !bytes.Equal(codes, header[2:]) {
			t.Fatalf("invalid codes field, expected: %v, got: %v", codes, header[2:])
		}
	})

	// other cases should be covered in fuzzing tests
}

func FuzzWriteHeader(f *testing.F) {
	testcases := []string{
		"Hello, world!",
		"",
		" ",
		"ab",
		"12345",
		"aaaaacccccbbbbbbbb",
		"aaacaaacccccbccbbaabbvccbbaab",
		"abcdefghijklmnopqrstuvwxyz",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		string([]byte{1, 23, 10, 11, 23, 75, 123, 233, 31, 255, 0, 11}),
	}
	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, a string) {
		b := []byte(a)
		writer := &bytes.Buffer{}
		reader := bytes.NewReader(b)
		encoder := NewEncoder(reader, writer)
		freq, err := getFrequencyMap(encoder.reader)
		if err != nil {
			t.Fatalf("unexpected error while computing frequency map on %v", b)
		}
		root := buildTree(freq)
		if err := encoder.writeHeader(root); err != nil {
			t.Fatalf("unexpected error while writinh header on %v, err: %v", b, err)
		}
		header := writer.Bytes()
		expectedLength := calculateTreeSize(root)
		length := binary.LittleEndian.Uint16(header)
		if expectedLength != length {
			t.Fatalf("invalid length writed, expected: %d, got: %d", expectedLength, length)
		}
		codes := func() []byte {
			buf := &bytes.Buffer{}
			writer := bitio.NewWriter(buf)
			if err := writeCodes(root, writer); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			writer.Flush()
			return buf.Bytes()
		}()
		if !bytes.Equal(codes, header[2:]) {
			t.Fatalf("invalid codes field, expected: %v, got: %v", codes, header[2:])
		}
	})
}

func TestEncodeContent(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		writer := &bytes.Buffer{}
		reader := bytes.NewReader([]byte{})
		encoder := NewEncoder(reader, writer)
		if err := encoder.encodeContent(map[byte]string{}, map[byte]uint{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		content := writer.Bytes()
		if len(content) != 8 {
			t.Fatalf("invalid length, expected 8, got: %v", len(content))
		}
		for _, b := range content {
			if b != 0 {
				t.Fatalf("unexpected bit in content length, expected all to be 0, got: %v", content)
			}
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		reader := bytes.NewReader([]byte("aaa"))
		encoder := NewEncoder(reader, &failingWriter{n: 3})
		codes := map[byte]string{'a': "0"}
		freq := map[byte]uint{'a': 3}
		if err := encoder.encodeContent(codes, freq); err == nil {
			t.Fatalf("expected writer error")
		}
	})

	// other cases should be covered in fuzzing tests
}

// func TestEncodeContentSingleByte(t *testing.T) {
// 	reader := bytes.NewReader([]byte{'A'})
// 	writer := &bytes.Buffer{}
// 	enc := NewEncoder(reader, writer)
// 	codes := map[byte]string{'a': "1"}
// 	size, _ := calculateContentSize(codes, map[byte]uint{'a': 1})

// 	if err := enc.encodeContent(codes, size); err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	data := writer.Bytes()
// 	if actualSize := binary.LittleEndian.Uint64(data); actualSize != size {
// 		t.Fatalf("invalid content size, expected %d, got %d", size, actualSize)
// 	}
// }

// func TestEncodeContentMultipleBytes(t *testing.T) {
// 	reader := bytes.NewReader([]byte{'a', 'b', 'a'})
// 	writer := &bytes.Buffer{}
// 	enc := NewEncoder(reader, writer)

// 	codes := map[byte]string{'a': "0", 'b': "1"}
// 	frequencies := map[byte]uint{'A': 2, 'B': 1}
// 	size, _ := calculateContentSize(codes, frequencies)

// 	if err := enc.encodeContent(codes, size); err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	data := writer.Bytes()
// 	if len(data) == 0 {
// 		t.Fatal("expected some bytes to be written")
// 	}
// }

// func TestEncodeContentVariedBitLengths(t *testing.T) {
// 	reader := bytes.NewReader([]byte{'a', 'b', 'c'})
// 	writer := &bytes.Buffer{}
// 	enc := NewEncoder(reader, writer)

// 	codes := map[byte]string{'A': "1", 'B': "01", 'C': "001"}

// 	if err := enc.encodeContent(codes, 1+2+3); err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	if writer.Len() == 0 {
// 		t.Fatal("expected data to be written")
// 	}
// }

// func TestEncodeContentWriterError(t *testing.T) {
// 	reader := bytes.NewReader([]byte{'A'})
// 	enc := NewEncoder(reader, &failingWriter{})

// 	codes := map[byte]string{'A': "1"}
// 	if err := enc.encodeContent(codes, 1); err == nil {
// 		t.Fatal("expected write error, got nil")
// 	}
// }

// func TestEncodeDeterministicLength(t *testing.T) {
// 	input := []byte("this is a test string for huffman encoder")
// 	repetitions := 5
// 	var lastLen int

// 	for i := range repetitions {
// 		r := bytes.NewReader(input)
// 		w := &bytes.Buffer{}
// 		enc := NewEncoder(r, w)

// 		if err := enc.Encode(); err != nil {
// 			t.Fatalf("Encode failed: %v", err)
// 		}

// 		encodedLen := w.Len()
// 		if i > 0 && encodedLen != lastLen {
// 			t.Fatalf("encoded length differs on repetition %d: got %d, want %d", i, encodedLen, lastLen)
// 		}

// 		lastLen = encodedLen
// 	}
// }
