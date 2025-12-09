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
	for i := range p {
		if err := e.WriteByte(p[i]); err != nil {
			return i, err
		}
	}
	return len(p), nil
}

func (e *errWriter) WriteByte(b byte) error {
	if e.limit <= 0 {
		return errors.New("write limit reached")
	}
	e.limit--
	return nil
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

func TestWriterByteByByte(t *testing.T) {
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

func TestWriterError(t *testing.T) {
	w := NewWriter(&errWriter{1})

	w.WriteBit(1)

	if _, err := w.Write([]byte{1, 2}); err != nil {
		if err := w.Flush(); err != nil {
			t.Fatalf("expected write error")
		}
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
