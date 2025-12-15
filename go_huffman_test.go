package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/serrhiy/go-huffman/benchkit"
)

type tempFiles struct {
	infile, outfile, resfile *os.File
}

func createTemporaryFiles() (*tempFiles, error) {
	infile, err := os.CreateTemp("", "go-huffman")
	if err != nil {
		return nil, err
	}
	outfile, err := os.CreateTemp("", "go-huffman")
	if err != nil {
		return nil, err
	}
	resFile, err := os.CreateTemp("", "go-huffman")
	if err != nil {
		return nil, err
	}
	return &tempFiles{infile, outfile, resFile}, nil
}

func encodeDecode(source string) (string, error) {
	files, err := createTemporaryFiles()
	if err != nil {
		return "", err
	}
	defer os.Remove(files.infile.Name())
	defer os.Remove(files.outfile.Name())
	defer os.Remove(files.resfile.Name())
	if _, err := bytes.NewBufferString(source).WriteTo(files.infile); err != nil {
		return "", err
	}
	if err := encodeFile(files.infile, files.outfile); err != nil {
		return "", err
	}
	files.outfile.Seek(0, io.SeekStart)
	if err := decodeFile(files.outfile, files.resfile); err != nil {
		return "", err
	}
	files.resfile.Seek(0, io.SeekStart)
	buffer, err := io.ReadAll(files.resfile)
	return string(buffer), err
}

func TestGoHuffman(t *testing.T) {
	testCases := []struct {
		name   string
		source string
	}{
		{
			name:   "empty",
			source: "",
		},
		{
			name:   "single char",
			source: "A",
		},
		{
			name:   "repeating char",
			source: "aaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			name:   "simple text",
			source: "hello world",
		},
		{
			name:   "multiline",
			source: "hello\nworld\nthis is huffman\n",
		},
		{
			name: "all bytes",
			source: func() string {
				b := make([]byte, 256)
				for i := range 256 {
					b[i] = byte(i)
				}
				return string(b)
			}(),
		},
		{
			name: "long text",
			source: func() string {
				b := make([]byte, 100_000)
				for i := range b {
					b[i] = byte('a' + i%26)
				}
				return string(b)
			}(),
		},
	}

	for _, tc := range testCases {
		result, err := encodeDecode(tc.source)
		if err != nil {
			t.Fatalf("%s encodeDecode failed: %v", tc.name, err)
		}
		if result != tc.source {
			t.Fatalf("%s failed, expected: '%s', got: '%s'", tc.name, tc.source, result)
		}
	}
}

func BenchmarkGoHuffman(b *testing.B) {
	cases := []struct {
		name string
		data string
	}{
		{"small_text", benchkit.Text(1 << 10)},
		{"medium_text", benchkit.Text(1 << 16)},
		{"large_text", benchkit.Text(1 << 20)},
		{"repeating", string(bytes.Repeat([]byte{'a'}, 1<<20))},
		{"random", benchkit.Random(1 << 20)},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				if _, err := encodeDecode(tc.data); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
