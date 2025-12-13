package bitio

import (
	"bytes"
	"errors"
	"math/rand"
	"testing"
)

type errWriter struct {
	limit int
}

func (e *errWriter) Write(p []byte) (int, error) {
	for i, b := range p {
		if err := e.WriteByte(b); err != nil {
			return i, err
		}
	}
	return len(p), nil
}

func (e *errWriter) WriteByte(byte) error {
	if e.limit <= 0 {
		return errors.New("write limit reached")
	}
	e.limit--
	return nil
}

func TestWriteByte(t *testing.T) {
	t.Run("write byte by byte", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)

		if err := w.WriteByte(0x12); err != nil {
			t.Fatalf("WriteByte error: %v", err)
		}
		if err := w.WriteByte(0x34); err != nil {
			t.Fatalf("WriteByte error: %v", err)
		}
		if err := w.Flush(); err != nil {
			t.Fatalf("WriteByte error: %v", err)
		}

		got := buf.Bytes()
		want := []byte{0x12, 0x34}

		if !bytes.Equal(got, want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})

	t.Run("write all bytes", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		for i := range 256 {
			if err := w.WriteByte(byte(i)); err != nil {
				t.Fatalf("unexpected error while writing byte: %v", err)
			}
		}
		w.Flush()
		values := buf.Bytes()
		for i := range 256 {
			if values[i] != byte(i) {
				t.Fatalf("invalid byte writed, expected: %d, got: %d", i, values[i])
			}
		}
	})

	t.Run("error propagation test", func(t *testing.T) {
		w := NewWriter(&errWriter{9})
		for i := range 9 {
			if err := w.WriteByte(byte(i)); err != nil {
				t.Fatalf("unexpected error while writing byte: %v", err)
			}
		}
		if err := w.Flush(); err != nil {
			t.Fatalf("unexpected error while flushing: %v", err)
		}
		w.WriteByte(1)
		if err := w.Flush(); err == nil {
			t.Fatalf("expected error, got: <nil>")
		}
	})
}

func TestWrite(t *testing.T) {
	t.Run("write 1 byte", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		w.Write([]byte{0xff})
		w.Flush()
		if buf.Len() != 1 || buf.Bytes()[0] != 0xff {
			t.Fatalf("invalud byte writed, expected: [0xff], got: %v", buf.Bytes())
		}
	})

	t.Run("write whole buffer", func(t *testing.T) {
		source := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		if _, err := w.Write(source); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := w.Flush(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(buf.Bytes(), source) {
			t.Fatalf("invalid bytes writed to buffer, expected: %v, got: %v", source, buf.Bytes())
		}
	})

	t.Run("write chunks", func(t *testing.T) {
		source := [][]byte{
			{1, 2, 3, 4},
			{5, 6, 7, 8},
			{9, 10, 11, 12},
		}
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		flatten := make([]byte, 0, len(source)*len(source[0]))
		for _, chunk := range source {
			if _, err := w.Write(chunk); err != nil {
				t.Fatalf("unexpexted error while writing chunk: %v", err)
			}
			flatten = append(flatten, chunk...)
		}
		if err := w.Flush(); err != nil {
			t.Fatalf("unexpexed error while flushing: %v", err)
		}
		if !bytes.Equal(flatten, buf.Bytes()) {
			t.Fatalf("invalid bytes writed, expected: %v, got: %v", flatten, buf.Bytes())
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		source := []byte{9: 0xff}
		w := NewWriter(&errWriter{9})
		w.Write(source)
		if err := w.Flush(); err == nil {
			t.Fatalf("error expected, got: <nil>")
		}
	})
}

func TestWriteBits(t *testing.T) {
	t.Run("write 1 bit(bit = 1)", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		w.WriteBits(0b10010101, 1)
		w.Flush()
		if buf.Bytes()[0] != (1 << 7) {
			t.Fatalf("invalid bit writed, expected: 1, got: 0")
		}
	})

	t.Run("write 1 bit(bit = 0)", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		w.WriteBits(0, 1)
		w.Flush()
		if buf.Bytes()[0] != 0 {
			t.Fatalf("invalid bit writed, expected: 0, got: 1")
		}
	})

	t.Run("write whole byte", func(t *testing.T) {
		const expected = 0b1100010
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		w.WriteBits(expected, 8)
		if err := w.Flush(); err != nil {
			t.Fatalf("unexpexted error whiule writing bits: %v", err)
		}
		if buf.Bytes()[0] != expected {
			t.Fatalf("invalid bits writed, expected: %#08b, got: %#08b", expected, buf.Bytes()[0])
		}
	})

	t.Run("write zero bytes", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		w.WriteBits(0xff, 0)
		w.Flush()
		if buf.Len() > 0 {
			t.Fatalf("invalid buffer length when writing 0 bits, expected: 0, got: %d", buf.Len())
		}
	})

	t.Run("invalud bytes length", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		if err := w.WriteBits(0xff, 9); err == nil {
			t.Fatal("expexted error when writing 9 bits to byte, got <nil>")
		}
	})

	t.Run("write part of byte", func(t *testing.T) {
		// brute force
		for i := range 256 {
			expected := byte(i)
			for j := range 9 {
				buf := &bytes.Buffer{}
				w := NewWriter(buf)
				w.WriteBits(expected, byte(j))
				if err := w.Flush(); err != nil {
					t.Fatalf("unexpected error while writing bits: %v", err)
				}
				if j == 0 {
					if buf.Len() != 0 {
						t.Fatalf("invalid buffer length when writing 0 bits, expected: 0, got: %d", buf.Len())
					}
					continue
				}
				mask := byte(((1 << j) - 1) << (8 - j))
				result := buf.Bytes()[0]
				if (expected & mask) != result {
					t.Fatalf("invalid bytes writed: expected: %d, got: %d", (expected & mask), result)
				}
			}
		}
	})
}

func TestWriterBasic(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)

	if err := w.WriteByte(0xC1); err != nil {
		t.Fatalf("WriteByte error: %v", err)
	}

	if err := w.WriteBit(1); err != nil {
		t.Fatalf("WriteBit errr: %v", err)
	}
	if err := w.WriteBit(0); err != nil {
		t.Fatalf("WriteBit errr: %v", err)
	}
	if err := w.WriteBit(1); err != nil {
		t.Fatalf("WriteBit errr: %v", err)
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("Flush error: %v", err)
	}

	if out := buf.Bytes(); len(out) == 0 {
		t.Fatalf("writer produced no output")
	}
}

func TestWriteBitPatterns(t *testing.T) {
	tests := []struct {
		bits []byte
		want byte
	}{
		{[]byte{0, 0, 0, 0, 0, 0, 0, 0}, 0x00},
		{[]byte{1, 1, 1, 1, 1, 1, 1, 1}, 0xFF},
		{[]byte{1, 0, 1, 0, 1, 0, 1, 0}, 0xAA},
		{[]byte{0, 1, 0, 1, 0, 1, 0, 1}, 0x55},
	}

	for _, tc := range tests {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)

		for _, b := range tc.bits {
			if err := w.WriteBit(b); err != nil {
				t.Fatalf("write error: %v", err)
			}
		}

		if err := w.Flush(); err != nil {
			t.Fatalf("flush error: %v", err)
		}

		if buf.Bytes()[0] != tc.want {
			t.Fatalf("expected %08b got %08b", tc.want, buf.Bytes()[0])
		}
	}
}

func TestWriterBitsAcrossByteBoundary(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)

	if err := w.WriteBit(1); err != nil {
		t.Fatalf("WriteBit err: %v", err)
	}
	if err := w.WriteBit(0); err != nil {
		t.Fatalf("WriteBit err: %v", err)
	}
	if err := w.WriteBit(1); err != nil {
		t.Fatalf("WriteBit err: %v", err)
	}

	if err := w.WriteByte(0xA5); err != nil {
		t.Fatalf("WriteByte err: %v", err)
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("Flush err: %v", err)
	}

	if len(buf.Bytes()) != 2 {
		t.Fatalf("expected 2 bytes, got %d", len(buf.Bytes()))
	}
}

func TestWriteBitsAcrossByte(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)

	if err := w.WriteBits(0b11100000, 3); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteBits(0b10101000, 5); err != nil {
		t.Fatal(err)
	}

	w.Flush()

	want := byte(0b11110101)
	got := buf.Bytes()[0]

	if got != want {
		t.Fatalf("expected %08b got %08b", want, got)
	}
}

func TestAlignWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)

	w.WriteBits(0b11100000, 3)

	if err := w.Align(); err != nil {
		t.Fatalf("Align error: %v", err)
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("Flush error: %v", err)
	}

	if len(buf.Bytes()) != 1 {
		t.Fatalf("expected 1 byte after align, got %d", len(buf.Bytes()))
	}

	want := byte(0b11100000)
	got := buf.Bytes()[0]

	if got != want {
		t.Fatalf("expected %08b got %08b", want, got)
	}

	if w.cacheSize != 0 {
		t.Fatalf("expected cacheSize=0 got %d", w.cacheSize)
	}
}

func TestReaderWriterChain(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)

	rng := rand.New(rand.NewSource(1))

	values := make([]byte, 50000)
	bits := make([]byte, len(values))

	for i := range values {
		values[i] = byte(rng.Int31n(255))
		bits[i] = byte(1 + rng.Int31n(7))
		err := w.WriteBits(values[i], bits[i])
		if err != nil {
			t.Fatalf("write error: %v", err)
		}
	}

	w.Flush()

	r := NewReader(bytes.NewBuffer(buf.Bytes()))

	for i := range values {
		var out byte
		var err error

		for b := byte(0); b < bits[i]; b++ {
			var bit byte
			bit, err = r.ReadBit()
			if err != nil {
				t.Fatalf("read error at %d: %v", i, err)
			}
			out = (out << 1) | bit
		}

		expect := values[i] >> (8 - bits[i])

		if out != expect {
			t.Fatalf("mismatch at %d: expected %08b got %08b", i, expect, out)
		}
	}
}

func TestAlignVariousCacheSizes(t *testing.T) {
	for bits := byte(1); bits < 8; bits++ {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)

		w.WriteBits(0xFF, bits)

		if err := w.Align(); err != nil {
			t.Fatalf("Align failed for bits=%d: %v", bits, err)
		}

		if err := w.Flush(); err != nil {
			t.Fatalf("Flush failes: %v", err)
		}

		if len(buf.Bytes()) != 1 {
			t.Fatalf("expected 1 byte for bits=%d got %d bytes", bits, len(buf.Bytes()))
		}

		result := buf.Bytes()[0]
		expected := byte(((1 << bits) - 1) << (8 - bits))

		if result != expected {
			t.Fatalf("invalid bit value: %d, expected: %d", result, expected)
		}
	}
}
