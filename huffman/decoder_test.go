package huffman

import (
	"bytes"
	"io"

	"encoding/binary"
	"errors"
	"testing"

	"github.com/serrhiy/go-huffman/bitio"
)

type brokenReader struct {
	limit  int
	reader io.ReadSeeker
}

func (r *brokenReader) Read(p []byte) (int, error) {
	if r.limit <= 0 {
		return 0, errors.New("read error")
	}

	if len(p) > r.limit {
		p = p[:r.limit]
	}

	n, err := r.reader.Read(p)
	r.limit -= n
	return n, err
}

func (b *brokenReader) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

func TestReadTree(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		reader := bitio.NewReader(&bytes.Buffer{})
		root, err := readTree(reader)
		if err != io.EOF {
			t.Fatalf("eof error expected, got: %v", err)
		}
		if root != nil {
			t.Fatalf("<nil> tree expected, got: %v", root)
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		source := []byte{10, 0, 0b01110000, 0b01000000}
		broken := &brokenReader{limit: 2, reader: bytes.NewReader(source)}
		root, err := readTree(bitio.NewReader(broken))
		if err == nil {
			t.Fatal("expected error, got: <nil>")
		}
		if root != nil {
			t.Fatalf("expected <nil> tree, got: %v", root)
		}
	})

	t.Run("1 leaf", func(t *testing.T) {
		source := []byte{10, 0, 0b01011000, 0b01000000}
		reader := bytes.NewReader(source)
		root, err := readTree(bitio.NewReader(reader))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root.isLeaf() {
			t.Fatal("root node canot be leaf")
		}
		if !root.left.isLeaf() || root.left.char != 'a' {
			t.Fatalf("invalid header reading, expected: 'a', got: %v", root.left)
		}
	})

	t.Run("2 leafs", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := bitio.NewWriter(buf)

		const size = 1 + 8 + 1 + 8 + 1
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, size)

		writer.Write(b)
		writer.WriteBit(0)
		writer.WriteBit(1)
		writer.WriteByte('A')
		writer.WriteBit(1)
		writer.WriteByte('B')
		writer.Flush()

		root, err := readTree(bitio.NewReader(buf))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if root.isLeaf() {
			t.Fatal("expected root to be internal node")
		}
		if root.left == nil || !root.left.isLeaf() || root.left.char != 'A' {
			t.Fatalf("left child incorrect: %+v", root.left)
		}
		if root.right == nil || !root.right.isLeaf() || root.right.char != 'B' {
			t.Fatalf("right child incorrect: %+v", root.right)
		}
	})

	// other cases should be covered in fuzzing tests
}

func TestDecode(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})

		decoder := NewDecoder(reader, nil)
		if err := decoder.Decode(); err == nil {
			t.Fatal("expected error when decoding empty file, got <nil>")
		}
	})

	t.Run("0 chars", func(t *testing.T) {
		reader := bytes.NewReader([]byte{9: 0})
		writer := &bytes.Buffer{}
		decoder := NewDecoder(reader, writer)
		if err := decoder.Decode(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if writer.Len() > 0 {
			t.Fatalf("expected empty result, got: %d", writer.Len())
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		source := []byte{10, 0, 0b01011000, 0b01000000, 1, 0, 0, 0, 0, 0, 0, 0, 0b10000000}
		reader := &brokenReader{limit: 12, reader: bytes.NewReader(source)}
		writer := &bytes.Buffer{}
		decoder := NewDecoder(reader, writer)
		if err := decoder.Decode(); err == nil {
			t.Fatal("error expected, got: <nil>")
		}
	})

	t.Run("decode 1 char", func(t *testing.T) {
		source := []byte{10, 0, 0b01011000, 0b01000000, 1, 0, 0, 0, 0, 0, 0, 0, 0b10000000}
		reader := bytes.NewReader(source)
		writer := &bytes.Buffer{}
		decoder := NewDecoder(reader, writer)
		if err := decoder.Decode(); err != nil {
			t.Fatalf("unexpected error while decoding: %v", err)
		}
		if writer.Bytes()[0] != 'a' {
			t.Fatalf("invalid decoder, expected: 'a', got: %q", writer.Bytes()[0])
		}
	})

	// other cases should be covered in fuzzing tests
}

func FuzzEncodeDecode(f *testing.F) {
	testcases := []string{
		"Hello world!",
		"",
		" ",
		"aaaaaaaaaaaaaaaaaaaaaaa",
		"abcdefghijklmnopqrstuvwxyz",
		"aaaaaaaabbbbbbbbbbbbbbbbb",
		string([]byte{0, 234, 14, 45, 13, 78, 32, 14}),
		"Duis quis quam sit amet diam semper congue. Donec ac auctor lectus",
		string(bytes.Repeat([]byte("abc"), 4096)),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, a string) {
		b := []byte(a)
		reader := bytes.NewReader(b)
		writer := &bytes.Buffer{}
		encoder := NewEncoder(reader, writer)
		if err := encoder.Encode(); err != nil {
			t.Fatalf("unexpected error while encoding: %v", err)
		}
		reader = bytes.NewReader(writer.Bytes())
		writer = &bytes.Buffer{}
		decoder := NewDecoder(reader, writer)
		if err := decoder.Decode(); err != nil {
			t.Fatalf("unexpected error while decoding: %v, input: %v", err, b)
		}
		if !bytes.Equal(b, writer.Bytes()) {
			t.Fatalf("invalid encoding decoding, input: %v, output: %v", b, writer.Bytes())
		}
	})
}
