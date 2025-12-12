package huffman

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/serrhiy/go-huffman/bitio"
)

func encodeDecode(input []byte, t *testing.T) []byte {
	in := bytes.NewReader(input)
	encoded := &bytes.Buffer{}
	decoded := &bytes.Buffer{}

	encoder := NewEncoder(in, encoded)
	if err := encoder.Encode(); err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	decoder := NewDecoder(bytes.NewReader(encoded.Bytes()), decoded)
	if err := decoder.Decode(); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	return decoded.Bytes()
}

func TestReadTreeSingleLeaf(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := bitio.NewWriter(buf)

	const size = 1 + 1 + 8
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, size)

	writer.Write(b)
	writer.WriteBit(0)
	writer.WriteBit(1)
	writer.WriteByte('A')
	writer.Flush()

	decoder := NewDecoder(buf, &bytes.Buffer{})
	root, err := decoder.readTree(bitio.NewReader(buf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root == nil || root.isLeaf() {
		t.Fatalf("expected inner node, got %+v", root)
	}
	if root.left == nil || !root.left.isLeaf() || root.left.char != 'A' {
		t.Fatalf("expected leaf node, got %+v", root.left)
	}
}

func TestReadTreeTwoLeaves(t *testing.T) {
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

	decoder := NewDecoder(buf, &bytes.Buffer{})
	root, err := decoder.readTree(bitio.NewReader(buf))
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
}

func TestReadTreeEmptyInput(t *testing.T) {
	decoder := NewDecoder(&bytes.Buffer{}, &bytes.Buffer{})
	reader := bitio.NewReader(decoder.reader)
	if _, err := decoder.readTree(reader); err != nil && err != io.EOF {
		t.Fatalf("reading an empty file should not cause an error: %v", err)
	}
}

func TestReadTreeEmptyReaderError(t *testing.T) {
	decoder := NewDecoder(&brokenReader{}, &bytes.Buffer{})
	reader := bitio.NewReader(decoder.reader)
	if _, err := decoder.readTree(reader); err == nil {
		t.Fatal("expected read error")
	}
}

func TestReadTreeLargeTree(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := bitio.NewWriter(buf)

	const size = 29

	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, size)
	writer.Write(b)

	// 		 (*)
	// 		/   \
	// 	(*)    C
	// 	/   \
	// A     B

	writer.WriteBit(0)
	writer.WriteBit(0)
	writer.WriteBit(1)
	writer.WriteByte('A')
	writer.WriteBit(1)
	writer.WriteByte('B')
	writer.WriteBit(1)
	writer.WriteByte('C')

	writer.Flush()

	decoder := NewDecoder(buf, &bytes.Buffer{})
	reader := bitio.NewReader(buf)

	root, err := decoder.readTree(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.left == nil || root.right == nil {
		t.Fatal("expected root to have two children")
	}
	if !root.left.left.isLeaf() || root.left.left.char != 'A' {
		t.Fatalf("left-left incorrect: %+v", root.left.left)
	}
	if !root.left.right.isLeaf() || root.left.right.char != 'B' {
		t.Fatalf("left-right incorrect: %+v", root.left.right)
	}
	if !root.right.isLeaf() || root.right.char != 'C' {
		t.Fatalf("right incorrect: %+v", root.right)
	}
}

func TestDecodeEmptyInputWritesNothing(t *testing.T) {
	out := &bytes.Buffer{}
	dec := NewDecoder(&bytes.Buffer{}, out)

	if err := dec.Decode(); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("expected empty output, got %v bytes", out.Len())
	}
}

func TestEncodeDecodeEmpty(t *testing.T) {
	out := encodeDecode([]byte{}, t)
	if len(out) != 0 {
		t.Fatalf("expected empty output, got %v bytes", len(out))
	}
}

func TestEncodeDecodeSingleCharacter(t *testing.T) {
	out := encodeDecode([]byte("A"), t)
	if string(out) != "A" {
		t.Fatalf("expected A, got %q", string(out))
	}
}

func TestEncodeDecodeRepeatedCharacter(t *testing.T) {
	out := encodeDecode([]byte("AAAAAAAAAAAA"), t)
	if string(out) != "AAAAAAAAAAAA" {
		t.Fatalf("expected repeated A, got %q", string(out))
	}
}

func TestEncodeDecodeTwoCharacters(t *testing.T) {
	out := encodeDecode([]byte("ABABABABAB"), t)
	if string(out) != "ABABABABAB" {
		t.Fatalf("decoded wrong: %q", string(out))
	}
}

func TestEncodeDecodeAlphabet(t *testing.T) {
	in := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	out := encodeDecode(in, t)
	if string(out) != string(in) {
		t.Fatalf("expected %q, got %q", string(in), string(out))
	}
}

func TestEncodeDecodeTextParagraph(t *testing.T) {
	in := []byte(`
Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`)
	out := encodeDecode(in, t)
	if !bytes.Equal(out, in) {
		t.Fatalf("decoded text mismatch")
	}
}

func TestEncodeDecodeBinaryData(t *testing.T) {
	in := []byte{0, 1, 2, 3, 4, 5, 10, 200, 255, 128, 64}
	out := encodeDecode(in, t)
	if !bytes.Equal(out, in) {
		t.Fatalf("decoded binary mismatch: %v != %v", out, in)
	}
}

func TestEncodeDecodeAllByteValues(t *testing.T) {
	in := make([]byte, 256)
	for i := range 256 {
		in[i] = byte(i)
	}
	out := encodeDecode(in, t)
	if !bytes.Equal(out, in) {
		t.Fatalf("decoded mismatch for all byte values")
	}
}

func TestEncodeDecodeSkewedFrequencies(t *testing.T) {
	in := []byte{}

	for range 10000 {
		in = append(in, 'A')
	}
	for range 10 {
		in = append(in, 'B')
	}

	out := encodeDecode(in, t)
	if !bytes.Equal(out, in) {
		t.Fatalf("decoded skewed mismatch")
	}
}

func TestEncodeDecodeDeepTree(t *testing.T) {
	in := []byte{}
	for i := range 200 {
		in = append(in, byte(i%250))
	}

	out := encodeDecode(in, t)
	if !bytes.Equal(out, in) {
		t.Fatalf("decoded deep-tree mismatch")
	}
}
