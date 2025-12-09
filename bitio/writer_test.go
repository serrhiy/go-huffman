package bitio

import (
	"bytes"
	"errors"
	"testing"
	"math/rand"
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

func TestWriterAlign(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)

	w.WriteBit(1)
	w.WriteBit(0)

	err := w.Align()
	if err != nil {
		t.Fatalf("Align failed: %v", err)
	}

	w.Flush()

	if len(buf.Bytes()) != 1 {
		t.Fatalf("expected 1 byte after align, got %d", len(buf.Bytes()))
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
