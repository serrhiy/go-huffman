package bitio

import (
	"bytes"
	"io"
	"testing"
)

func TestReadByte(t *testing.T) {
	t.Run("empty buffer", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{}))
		if _, err := r.ReadByte(); err != io.EOF {
			t.Fatalf("error expected while reading empty buffer, got: %v", err)
		}
	})

	t.Run("read 1 byte", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0xff}))
		res, err := r.ReadByte()
		if err != nil {
			t.Fatalf("unexpected error while reding byte: %v", err)
		}
		if res != 0xff {
			t.Fatalf("invalid byte readed, expected: %d, got: %d", 0xff, res)
		}
	})

	t.Run("read several bytes", func(t *testing.T) {
		source := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
		r := NewReader(bytes.NewBuffer(source))
		buffer := make([]byte, 0, len(source))
		for len(buffer) < len(source) {
			res, err := r.ReadByte()
			if err != nil {
				t.Fatalf("unxepexted error while reading byte: %v", err)
			}
			buffer = append(buffer, res)
		}
		if !bytes.Equal(source, buffer) {
			t.Fatalf("original and read buffers do not match: expexted: %v, got: %v", source, buffer)
		}
	})

	t.Run("EOF after reading whole buffer", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{1, 2}))
		r.ReadByte()
		r.ReadByte()
		if _, err := r.ReadByte(); err != io.EOF {
			t.Fatalf("EOF expected after reading whole buffer: %v", err)
		}
	})
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

	t.Run("invalid number of bytes to read", func(t *testing.T) {
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
		if res != 1<<7 {
			t.Fatalf("nvalid reasult of reading 1 bit, expected: %d, got: %d", 1<<7, res)
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

func TestRead(t *testing.T) {
	t.Run("empty buffer", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{}))
		buffer := make([]byte, 1)
		n, err := r.Read(buffer)
		if err != io.EOF {
			t.Fatalf("expexted error while reading from empty buffer, got: %d, %v", n, err)
		}
		if n != 0 {
			t.Fatalf("readed more than zero bytes from empty buffer: %d", n)
		}
	})

	t.Run("read zero bytes", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{1, 2, 3}))
		buffer := []byte{}
		n, err := r.Read(buffer)
		if err != nil {
			t.Fatalf("unexpected error while reading 0 bytes: %v", err)
		}
		if n != 0 {
			t.Fatalf("readed more than zero bytes to empty buffer: %d", n)
		}
	})

	t.Run("read zero bytes from empty buffer", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{}))
		buffer := []byte{}
		n, err := r.Read(buffer)
		if err != nil {
			t.Fatalf("unexpected error while reading 0 bytes from empty buffer: %v", err)
		}
		if n != 0 {
			t.Fatalf("readed more than zero bytes from empty buffer: %d", n)
		}
	})

	t.Run("read 1 byte", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{10}))
		buffer := make([]byte, 2)
		n, err := r.Read(buffer)
		if err != nil {
			t.Fatalf("unxepected error while reading 1 byte: %v", err)
		}
		if n != 1 {
			t.Fatalf("invalid number of readed bytes, expected: 1, got: %v", n)
		}
	})

	t.Run("read multiple bytes", func(t *testing.T) {
		source := []byte{10, 20, 30, 40}
		r := NewReader(bytes.NewBuffer(source))
		buffer := make([]byte, len(source))
		n, err := r.Read(buffer)
		if err != nil {
			t.Fatalf("unexpected error while reading whole buffer: %v", err)
		}
		if n != len(source) {
			t.Fatalf("invalid number of readed bytes, expected: %d, got: %d", len(source), n)
		}
		if !bytes.Equal(source, buffer) {
			t.Fatalf("invalid bytes readed from buffer, expected: %v, got: %v", source, buffer)
		}
	})

	t.Run("partial buffer read", func(t *testing.T) {
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
	})

	t.Run("read whole buffer by parts", func(t *testing.T) {
		source := []byte{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
		r := NewReader(bytes.NewBuffer(source))
		buffer := make([]byte, 0, len(source))
		for i := 1; len(buffer) < len(source); i++ {
			toRead := min(i, len(source)-len(buffer))
			b := make([]byte, toRead)
			if _, err := r.Read(b); err != nil {
				t.Fatalf("unexpected error while reading bytes: %v", err)
			}
			buffer = append(buffer, b...)
		}
		if !bytes.Equal(source, buffer) {
			t.Fatalf("invalid bytes readed from buffer, expected: %v, got: %v", source, buffer)
		}
	})

	t.Run("read more than is in the buffer", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{1, 2, 3}))
		buffer := make([]byte, 4)
		n, _ := r.Read(buffer)
		if n != 3 {
			t.Fatalf("invalid number of readed bytes, expected: %d, got: %d", 3, n)
		}
		if _, err := r.Read(buffer); err != io.EOF {
			t.Fatalf("expected EOF, got: %v", err)
		}
	})

	t.Run("EOF after byte reading", func(t *testing.T) {
		source := []byte{10, 20, 30}
		r := NewReader(bytes.NewBuffer(source))
		buffer := make([]byte, 2)
		r.Read(buffer)
		if n, _ := r.Read(buffer); n != 1 {
			t.Fatalf("expected to read 1 byte, got: %d", n)
		}
		if _, err := r.Read(buffer); err != io.EOF {
			t.Fatalf("expected EOF, got: %v", err)
		}
	})
}

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

