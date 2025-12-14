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

}

// func encodeDecode(input []byte, t *testing.T) []byte {
// 	in := bytes.NewReader(input)
// 	encoded := &bytes.Buffer{}
// 	decoded := &bytes.Buffer{}

// 	encoder := NewEncoder(in, encoded)
// 	if err := encoder.Encode(); err != nil {
// 		t.Fatalf("encode failed: %v", err)
// 	}

// 	decoder := NewDecoder(bytes.NewReader(encoded.Bytes()), decoded)
// 	if err := decoder.Decode(); err != nil {
// 		t.Fatalf("decode failed: %v", err)
// 	}

// 	return decoded.Bytes()
// }

// func TestDecodeEmptyInputWritesNothing(t *testing.T) {
// 	out := &bytes.Buffer{}
// 	dec := NewDecoder(&bytes.Buffer{}, out)

// 	if err := dec.Decode(); err != nil {
// 		t.Fatalf("Decode failed: %v", err)
// 	}

// 	if out.Len() != 0 {
// 		t.Fatalf("expected empty output, got %v bytes", out.Len())
// 	}
// }

// func TestEncodeDecodeEmpty(t *testing.T) {
// 	out := encodeDecode([]byte{}, t)
// 	if len(out) != 0 {
// 		t.Fatalf("expected empty output, got %v bytes", len(out))
// 	}
// }

// func TestEncodeDecodeSingleCharacter(t *testing.T) {
// 	out := encodeDecode([]byte("A"), t)
// 	if string(out) != "A" {
// 		t.Fatalf("expected A, got %q", string(out))
// 	}
// }

// func TestEncodeDecodeRepeatedCharacter(t *testing.T) {
// 	out := encodeDecode([]byte("AAAAAAAAAAAA"), t)
// 	if string(out) != "AAAAAAAAAAAA" {
// 		t.Fatalf("expected repeated A, got %q", string(out))
// 	}
// }

// func TestEncodeDecodeTwoCharacters(t *testing.T) {
// 	out := encodeDecode([]byte("ABABABABAB"), t)
// 	if string(out) != "ABABABABAB" {
// 		t.Fatalf("decoded wrong: %q", string(out))
// 	}
// }

// func TestEncodeDecodeAlphabet(t *testing.T) {
// 	in := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
// 	out := encodeDecode(in, t)
// 	if string(out) != string(in) {
// 		t.Fatalf("expected %q, got %q", string(in), string(out))
// 	}
// }

// func TestEncodeDecodeTextParagraph(t *testing.T) {
// 	in := []byte(`
// Lorem ipsum dolor sit amet, consectetur adipiscing elit.
// Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
// `)
// 	out := encodeDecode(in, t)
// 	if !bytes.Equal(out, in) {
// 		t.Fatalf("decoded text mismatch")
// 	}
// }

// func TestEncodeDecodeBinaryData(t *testing.T) {
// 	in := []byte{0, 1, 2, 3, 4, 5, 10, 200, 255, 128, 64}
// 	out := encodeDecode(in, t)
// 	if !bytes.Equal(out, in) {
// 		t.Fatalf("decoded binary mismatch: %v != %v", out, in)
// 	}
// }

// func TestEncodeDecodeAllByteValues(t *testing.T) {
// 	in := make([]byte, 256)
// 	for i := range 256 {
// 		in[i] = byte(i)
// 	}
// 	out := encodeDecode(in, t)
// 	if !bytes.Equal(out, in) {
// 		t.Fatalf("decoded mismatch for all byte values")
// 	}
// }

// func TestEncodeDecodeSkewedFrequencies(t *testing.T) {
// 	in := []byte{}

// 	for range 10000 {
// 		in = append(in, 'A')
// 	}
// 	for range 10 {
// 		in = append(in, 'B')
// 	}

// 	out := encodeDecode(in, t)
// 	if !bytes.Equal(out, in) {
// 		t.Fatalf("decoded skewed mismatch")
// 	}
// }

// func TestEncodeDecodeDeepTree(t *testing.T) {
// 	in := []byte{}
// 	for i := range 200 {
// 		in = append(in, byte(i%250))
// 	}

// 	out := encodeDecode(in, t)
// 	if !bytes.Equal(out, in) {
// 		t.Fatalf("decoded deep-tree mismatch")
// 	}
// }
