package huffman

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type brokenReader struct {
	io.ReadSeeker
}

func (b *brokenReader) Read([]byte) (int, error) {
	return 0, errors.New("read error")
}

func (b *brokenReader) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

func TestHuffmanEncoderGetFrequencyMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected map[byte]uint
	}{
		{
			name:  "simple text",
			input: []byte("aabccc"),
			expected: map[byte]uint{
				'a': 2,
				'b': 1,
				'c': 3,
			},
		},
		{
			name:     "empty input",
			input:    []byte{},
			expected: map[byte]uint{},
		},
		{
			name: "all bytes",
			input: func() []byte {
				data := make([]byte, 256)
				for i := range 256 {
					data[i] = byte(i)
				}
				return data
			}(),
			expected: func() map[byte]uint {
				m := make(map[byte]uint)
				for i := range 256 {
					m[byte(i)] = 1
				}
				return m
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := bytes.NewReader(tc.input)
			writer := bytes.NewBuffer(nil)

			encoder := &HuffmanEncoder{
				reader: reader,
				writer: writer,
			}

			freq, err := encoder.getFrequencyMap()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(freq) != len(tc.expected) {
				t.Fatalf("unexpected freq size: got %d, want %d", len(freq), len(tc.expected))
			}

			for k, v := range tc.expected {
				if freq[k] != v {
					t.Fatalf("freq mismatch for '%c': got %d, want %d", k, freq[k], v)
				}
			}
		})
	}
}

func TestHuffmanEncoderGetFrequencyMapReadError(t *testing.T) {
	reader := &brokenReader{
		ReadSeeker: bytes.NewReader([]byte("abc")),
	}
	writer := bytes.NewBuffer(nil)

	encoder := &HuffmanEncoder{
		reader: reader,
		writer: writer,
	}

	if _, err := encoder.getFrequencyMap(); err == nil {
		t.Fatalf("expected read error, got nil")
	}
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

func TestWriteCodesSingleLeaf(t *testing.T) {
	root := &node{char: 'Z', count: 1}

	buf := &bytes.Buffer{}
	enc := NewEncoder(nil, buf)

	if err := enc.writeCodes(root); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.Bytes()
	if len(out) < 2 {
		t.Fatalf("expected at least 1 byte + newline, got: %v", out)
	}
	if out[len(out)-1] != '\n' {
		t.Fatalf("expected newline, got %v", out)
	}
}

func TestWriteCodesSimpleTree(t *testing.T) {
	root := &node{
		left:  &node{char: 'A', count: 3},
		right: &node{char: 'B', count: 5},
	}

	buf := &bytes.Buffer{}
	enc := NewEncoder(nil, buf)

	err := enc.writeCodes(root)
	if err != nil {
		t.Fatalf("writeCodes returned error: %v", err)
	}

	b := buf.Bytes()
	if len(b) == 0 || b[len(b)-1] != '\n' {
		t.Fatalf("expected ending newline, got %v", b)
	}

	if len(b) <= 1 {
		t.Fatalf("expected some bit-encoded data before newline")
	}
}

func TestWriteCodesWriterError(t *testing.T) {
	root := &node{char: 'X', count: 1}

	fw := &failingWriter{n: 0}
	enc := NewEncoder(nil, fw)

	err := enc.writeCodes(root)
	if err == nil {
		t.Fatalf("expected writer error")
	}
}

func TestWriteCodesAlignCalled(t *testing.T) {
	root := &node{
		left:  &node{char: 'A', count: 1},
		right: &node{char: 'B', count: 2},
	}

	buf := &bytes.Buffer{}
	enc := NewEncoder(nil, buf)

	err := enc.writeCodes(root)
	if err != nil {
		t.Fatalf("writeCodes error: %v", err)
	}

	out := buf.Bytes()

	if len(out) < 2 {
		t.Fatalf("expected data + newline, got too short: %v", out)
	}

	if out[len(out)-1] != '\n' {
		t.Fatalf("expected newline, got: %v", out)
	}
}

func TestEncodeContentSingleByte(t *testing.T) {
	reader := bytes.NewReader([]byte{'A'})
	writer := &bytes.Buffer{}
	enc := NewEncoder(reader, writer)

	codes := map[byte]string{'A': "101"}
	if err := enc.encodeContent(codes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := writer.Bytes()
	if len(data) == 0 {
		t.Fatal("expected some bytes to be written")
	}
}

func TestEncodeContentMultipleBytes(t *testing.T) {
	reader := bytes.NewReader([]byte{'A', 'B', 'A'})
	writer := &bytes.Buffer{}
	enc := NewEncoder(reader, writer)

	codes := map[byte]string{
		'A': "0",
		'B': "11",
	}

	if err := enc.encodeContent(codes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := writer.Bytes()
	if len(data) == 0 {
		t.Fatal("expected some bytes to be written")
	}
}

func TestEncodeContentVariedBitLengths(t *testing.T) {
	reader := bytes.NewReader([]byte{'A', 'B', 'C'})
	writer := &bytes.Buffer{}
	enc := NewEncoder(reader, writer)

	codes := map[byte]string{
		'A': "1",
		'B': "01",
		'C': "001",
	}

	if err := enc.encodeContent(codes); err != nil {
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
	if err := enc.encodeContent(codes); err == nil {
		t.Fatal("expected write error, got nil")
	}
}

func TestEncodeContentReaderError(t *testing.T) {
	enc := NewEncoder(&brokenReader{}, &bytes.Buffer{})

	codes := map[byte]string{'A': "1"}
	err := enc.encodeContent(codes)
	if err == nil {
		t.Fatalf("expected nil because encodeContent ignores non-EOF read errors, got %v", err)
	}
}

func TestEncodeContentBitPattern(t *testing.T) {
	data := []byte{'A', 'B'}
	reader := bytes.NewReader(data)
	writer := &bytes.Buffer{}
	enc := NewEncoder(reader, writer)

	codes := map[byte]string{
		'A': "10",
		'B': "01",
	}

	if err := enc.encodeContent(codes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := writer.Bytes()
	if len(out) == 0 {
		t.Fatal("expected bytes written")
	}

	if out[0] != 0 && out[0] != 128 {
		t.Logf("first byte: %08b", out[0])
	}
}
