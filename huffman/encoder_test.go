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

func TestEncode(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		writer := &bytes.Buffer{}
		reader := bytes.NewReader([]byte{})
		encoder := NewEncoder(reader, writer)
		if err := encoder.Encode(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result := writer.Bytes()
		if len(result) != 10 {
			t.Fatalf("after enc0ding empty buffer content size should be 10, actual: %d", len(result))
		}
		for _, b := range result {
			if b != 0 {
				t.Fatalf("all bytes should be 0, actual: %v", result)
			}
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		reader := bytes.NewReader([]byte("aaa"))
		fw := &failingWriter{n: 5}
		encoder := NewEncoder(reader, fw)

		err := encoder.Encode()
		if err == nil {
			t.Fatalf("expected writer error")
		}
	})

	t.Run("1 char", func(t *testing.T) {
		writer := &bytes.Buffer{}
		reader := bytes.NewReader([]byte{'a'})
		encoder := NewEncoder(reader, writer)
		if err := encoder.Encode(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		result := writer.Bytes()
		if len(result) != 13 {
			t.Fatalf("after encoding 1 byte content size should be 13, actual: %d", len(result))
		}
		headerSize := binary.LittleEndian.Uint16(result)
		if headerSize != 10 {
			t.Fatalf("invalid header size, expected: %d, got: %d", 10, headerSize)
		}
		if (result[2] & 0b11000000) != 0b01000000 {
			t.Fatalf("first two bits of header should be 01, got: %#08b", (result[2] & 0b11000000))
		}
		char := byte(((result[2] & 0b00111111) << 2) | (result[3]&0b11000000)>>6)
		if char != 'a' {
			t.Fatalf("invalid char writed, expected: 'a', got: %q", char)
		}
		if (result[3] & 0b00111111) != 0 {
			t.Fatalf("the rest of byte must consist of zeros, got: %#08b", (result[3] & 0b00111111))
		}
		contentSize := binary.LittleEndian.Uint64(result[4:])
		if contentSize != 1 {
			t.Fatalf("invalid content size, expected: %d, got: %d", 1, contentSize)
		}
		if (result[12]&0b10000000)>>7 != 1 {
			t.Fatalf("invalid content field, expected 1, got: %d", (result[12]&0b10000000)>>7)
		}
		if result[12]&0b01111111 != 0 {
			t.Fatalf("final padding bits must be zero-filled, got: %d", result[12]&0b01111111)
		}
	})

	// other cases should be covered in fuzzing tests
}
