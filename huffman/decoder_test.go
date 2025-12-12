package huffman

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/serrhiy/go-huffman/bitio"
)

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

func TestDecode_EmptyInputWritesNothing(t *testing.T) {
	out := &bytes.Buffer{}
	dec := NewDecoder(&bytes.Buffer{}, out)

	if err := dec.Decode(); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("expected empty output, got %v bytes", out.Len())
	}
}
