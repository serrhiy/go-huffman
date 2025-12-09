package bitio

import (
	"bytes"
	"io"
	"testing"
)

func TestReaderBasic(t *testing.T) {
	data := []byte{0xC1, 0b01000000}

	r := NewReader(bytes.NewBuffer(data))

	b, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte error: %v", err)
	}
	if b != 0xC1 {
		t.Fatalf("expected 0xC1, got %02x", b)
	}

	bit, err := r.ReadBit()
	if err != nil || bit != 0 {
		t.Fatalf("expected bit=1, got %d err=%v", bit, err)
	}

	bit, err = r.ReadBit()
	if err != nil || bit != 1 {
		t.Fatalf("expected bit=1, got %d err=%v", bit, err)
	}
}

func TestReaderBitByBitFullByte(t *testing.T) {
	r := NewReader(bytes.NewBuffer([]byte{0b10110010}))

	expected := []byte{1, 0, 1, 1, 0, 0, 1, 0}

	for i, exp := range expected {
		b, err := r.ReadBit()
		if err != nil {
			t.Fatalf("err at bit %d: %v", i, err)
		}
		if b != exp {
			t.Fatalf("bit %d: expected %d got %d", i, exp, b)
		}
	}
}

func TestReaderReadUnaligned(t *testing.T) {
	r := NewReader(bytes.NewBuffer([]byte{0b01010110, 0b10010110}))

	r.ReadBit()

	b, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte error: %v", err)
	}

	if b != 0b10101101 {
		t.Fatalf("expected 0x82, got %02x", b)
	}
}

func TestReaderReadEOF(t *testing.T) {
	r := NewReader(bytes.NewBuffer([]byte{0x01}))

	b, err := r.ReadByte()
	if err != nil || b != 0x01 {
		t.Fatalf("expected first byte OK")
	}

	_, err = r.ReadByte()
	if err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}

	_, err = r.ReadBit()
	if err != io.EOF {
		t.Fatalf("expected EOF on ReadBit, got: %v", err)
	}

	buf := make([]byte, 3)
	n, err := r.Read(buf)
	if n != 0 || err != io.EOF {
		t.Fatalf("expected EOF read, got n=%d err=%v", n, err)
	}
}

func TestReadBitAndByteBoundary(t *testing.T) {
	data := []byte{0b10110011, 0b01101100}
	r := NewReader(bytes.NewBuffer(data))

	var bits byte
	for range 3 {
		b, err := r.ReadBit()
		if err != nil {
			t.Fatalf("ReadBit error: %v", err)
		}
		bits = (bits << 1) | b
	}
	if bits != 0b101 {
		t.Fatalf("expected 101, got %03b", bits)
	}

	b, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte error: %v", err)
	}

	if b != 0b10011011 {
		t.Fatalf("expected 10011011, got %08b", b)
	}
}

func TestPartialBufferRead(t *testing.T) {
	data := []byte{0x11, 0x22, 0x33, 0x44}
	r := NewReader(bytes.NewBuffer(data))

	buf := make([]byte, 2)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if n != 2 || buf[0] != 0x11 || buf[1] != 0x22 {
		t.Fatalf("expected [0x11 0x22], got %v", buf[:n])
	}

	n, err = r.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read error: %v", err)
	}
	if n != 2 || buf[0] != 0x33 || buf[1] != 0x44 {
		t.Fatalf("expected [0x33 0x44], got %v", buf[:n])
	}
}
