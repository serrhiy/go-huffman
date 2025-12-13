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

func TestReaderReadByteUnaligned(t *testing.T) {
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

func TestReaderAlign(t *testing.T) {
	data := []byte{0b10101010, 0b11001100}
	r := NewReader(bytes.NewBuffer(data))

	for range 3 {
		if _, err := r.ReadBit(); err != nil {
			t.Fatalf("ReadBit error: %v", err)
		}
	}

	if err := r.Align(); err != nil {
		t.Fatalf("Align error: %v", err)
	}

	if r.cacheSize != 0 {
		t.Fatalf("Expected empty cache, got: %d, %d", r.cacheSize, r.cache)
	}

	b, err := r.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte error: %v", err)
	}

	if b != 0b11001100 {
		t.Fatalf("expected 11001100, got %08b", b)
	}
}

func TestReaderReadUnalignedEOF(t *testing.T) {
	r := NewReader(bytes.NewBuffer([]byte{0b11110000}))

	r.ReadBit()
	r.ReadBit()

	buf := make([]byte, 1)

	if _, err := r.Read(buf); err != io.EOF {
		t.Fatalf("expected EOF, got: %v", err)
	}
}

func TestReaderReadUnaligned(t *testing.T) {
	r := NewReader(bytes.NewBuffer([]byte{
		0b10110010,
		0b01011010,
		0b01110011,
	}))
	r.ReadBit()
	r.ReadBit()

	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf[0] != 0b11001001 {
		t.Fatalf("first bit is incorrect, expected: %d, got: %d", 0b11001001, buf[0])
	}
	if buf[1] != 0b01101001 {
		t.Fatalf("first bit is incorrect, expected: %d, got: %d", 0b01101001, buf[1])
	}
}

func TestReaderBitAfterUnalignedByte(t *testing.T) {
	r := NewReader(bytes.NewBuffer([]byte{0b11001010, 0b10000000}))

	r.ReadBit()
	r.ReadBit()

	b, _ := r.ReadByte()
	if b != 0b00101010 {
		t.Fatalf("unexpected byte: %08b", b)
	}

	bit, _ := r.ReadBit()
	if bit != 0 {
		t.Fatalf("expected next bit=0 got %d", bit)
	}
}

func TestReaderReadZeroBytes(t *testing.T) {
	r := NewReader(bytes.NewBuffer([]byte{0xAA}))

	n, err := r.Read([]byte{})
	if n != 0 || err != nil {
		t.Fatalf("expected (0, nil), got (%d, %v)", n, err)
	}
}

func TestReaderEmptyInput(t *testing.T) {
	r := NewReader(bytes.NewBuffer(nil))

	if _, err := r.ReadBit(); err != io.EOF {
		t.Fatalf("expected EOF for ReadBit, got %v", err)
	}

	if _, err := r.ReadByte(); err != io.EOF {
		t.Fatalf("expected EOF for ReadByte, got %v", err)
	}

	buf := make([]byte, 5)
	n, err := r.Read(buf)
	if n != 0 || err != io.EOF {
		t.Fatalf("expected EOF for Read(), got (%d, %v)", n, err)
	}
}

func TestReaderAlignOnBoundary(t *testing.T) {
	r := NewReader(bytes.NewBuffer([]byte{0xAA, 0xBB}))

	if err := r.Align(); err != nil {
		t.Fatalf("Align returned error on empty cache: %v", err)
	}

	b, err := r.ReadByte()
	if err != nil || b != 0xAA {
		t.Fatalf("expected 0xAA, got %02x err=%v", b, err)
	}

	if err := r.Align(); err != nil {
		t.Fatalf("Align should still work on boundary")
	}

	b, err = r.ReadByte()
	if err != nil || b != 0xBB {
		t.Fatalf("expected 0xBB, got %02x err=%v", b, err)
	}
}

func TestReadBits(t *testing.T) {
	t.Run("empty buffer", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{}))
		if res, err := r.ReadBits(1); err != io.EOF {
			t.Fatalf("EOF error expected, got: %d, %v", res, err)
		}
	})

	t.Run("read 0 bits", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{1, 2, 3}))
		res, err := r.ReadBits(0)
		if err != nil {
			t.Fatalf("unexpected error while reading 0 bits: %v", err)
		}
		if res != 0 {
			t.Fatalf("invalid result after reading 0 bits: %d", res)
		}
	})

	t.Run("invalid number of bytes to read", func (t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{1, 2, 3}))
		if _, err := r.ReadBits(9); err == nil {
			t.Fatal("no error when reading > 8 bits")
		}
	})

	t.Run("EOF", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0b11001010}))
		r.ReadBit()
		_, err := r.ReadBits(8)
		if err != io.EOF {
			t.Fatalf("expected EOF but got: %v", err)
		}
	})

	t.Run("reading 1 bit(bit = 1)", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0b11001010}))
		res, err := r.ReadBits(1)
		if err != nil {
			t.Fatalf("unexpexted error while reading 1 bit: %v", err)
		}
		if res != 1 << 7 {
			t.Fatalf("nvalid reasult of reading 1 bit, expected: %d, got: %d", 1 << 7, res)
		}
	})

	t.Run("reading 1 bit(bit = 0)", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0b01001010}))
		res, err := r.ReadBits(1)
		if err != nil {
			t.Fatalf("unexpexted error while reading 1 bit: %v", err)
		}
		if res != 0 {
			t.Fatalf("invalid reasult of reading 1 bit, expected: %d, got: %d", 0, res)
		}
	})

	t.Run("reading 8 bits", func(t *testing.T) {
		const expected = 0b01001010
		r := NewReader(bytes.NewBuffer([]byte{0b01001010}))
		res, err := r.ReadBits(8)
		if err != nil {
			t.Fatalf("unexpected error while reading 8 bits: %v", err)
		}
		if res != expected {
			t.Fatalf("invalid reasult of reading 8 bits, expected: %d, got: %d", expected, res)
		}
	})

	t.Run("reading several bits inside byte", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0b01001010}))
		res, err := r.ReadBits(3)
		if err != nil {
			t.Fatalf("unexpexted error while reading bits: %v", err)
		}
		if (res & 0b11100000) != 0b01000000 {
			t.Fatalf("invalud reading result, expected: %d, got: %d", 0b01000000, res)
		}
		res, err = r.ReadBits(5)
		if err != nil {
			t.Fatalf("unexpexted error while reading bits: %v", err)
		}
		if (res & 0b11111000) != 0b01010000 {
			t.Fatalf("invalud reading result, expected: %d, got: %d", 0b01010000, res)
		}

		if r.cache != 0 {
			t.Fatalf("invalid cache value, expected: 0, result: %d", r.cache)
		}

		if r.cacheSize != 0 {
			t.Fatalf("invalid cache size value, expected: 0, result: %d", r.cacheSize)
		}
	})

	t.Run("read bits across boundary", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0b01001010, 0b10100101}))
		r.ReadBits(5)
		res, err := r.ReadBits(8)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != 0b01010100 {
			t.Fatalf("invaliud result value, expected: %v, got: %v", 0b01010100, res)
		}
	})
}