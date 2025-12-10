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
