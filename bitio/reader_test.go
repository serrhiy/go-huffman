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

func TestReadBit(t *testing.T) {
	t.Run("empty buffer", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{}))
		if res, err := r.ReadBit(); err != io.EOF {
			t.Fatalf("EOF expected when reading empty file, got: %d, %v", res, err)
		}
	})

	t.Run("read 1 bit(bit = 1)", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0b10010101}))
		bit, err := r.ReadBit()
		if err != nil {
			t.Fatalf("unexpected error while reading 1 bit: %v", err)
		}
		if bit != 1 {
			t.Fatalf("invalid bit readed, expected 1, got: %d", bit)
		}
	})

	t.Run("read 1 bit(bit = 0)", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0b01010101}))
		bit, err := r.ReadBit()
		if err != nil {
			t.Fatalf("unexpected error while reading 1 bit: %v", err)
		}
		if bit != 0 {
			t.Fatalf("invalid bit readed, expected 0, got: %d", bit)
		}
	})

	t.Run("read all bits inside byte", func(t *testing.T) {
		const expected = 0b10010101
		r := NewReader(bytes.NewBuffer([]byte{expected}))
		var result byte = 0
		for i := 1; i <= 8; i++ {
			bit, err := r.ReadBit()
			if err != nil {
				t.Fatalf("unexpected error occured while reading bit: %v", err)
			}
			result |= bit << (8 - i)
		}
		if result != expected {
			t.Fatalf("invalid byte after reading 8 bits, expected: %#b, got: %#b", expected, result)
		}
		if r.cache != 0 || r.cacheSize != 0 {
			t.Fatalf("after reading all bits cache should be empty, cache: %d, size: %d", r.cache, r.cacheSize)
		}
	})

	t.Run("read several bytes", func(t *testing.T) {
		source := []byte{0b10110101, 0b00100101, 0b01001000, 0b10101010, 0b00110001}
		r := NewReader(bytes.NewBuffer(source))
		buffer := make([]byte, 0, len(source))
		for len(buffer) < len(source) {
			var result byte = 0
			for i := 1; i <= 8; i++ {
				bit, err := r.ReadBit()
				if err != nil {
					t.Fatalf("unexpected error occured while reading bit: %v", err)
				}
				result |= bit << (8 - i)
			}
			buffer = append(buffer, result)
		}
		if !bytes.Equal(source, buffer) {
			t.Fatalf("invalud bytes readed, expected: %v, got: %v", source, buffer)
		}
		if r.cache != 0 || r.cacheSize != 0 {
			t.Fatalf("after reading all bits cache should be empty, cache: %d, size: %d", r.cache, r.cacheSize)
		}
	})

	t.Run("EOF after reading all bits", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0xff}))
		for range 8 {
			r.ReadBit()
		}
		if _, err := r.ReadBit(); err != io.EOF {
			t.Fatalf("eof expected after reading all bits, got: %v", err)
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

func TestAlign(t *testing.T) {
	t.Run("empty buffer", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{}))
		if err := r.Align(); err != nil {
			t.Fatalf("unexpected error while aligning empty buffer: %v", err)
		}
	})

	t.Run("align after reading 1 bit", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0xff}))
		r.ReadBit()
		if err := r.Align(); err != nil {
			t.Fatalf("unexpected error while aligning: %v", err)
		}
		if r.cache != 0 || r.cacheSize != 0 {
			t.Fatalf("cache and cache size should be empty after aligning, cache: %d, cache size: %d", r.cache, r.cacheSize)
		}
	})

	t.Run("align different size", func(t *testing.T) {
		for i := range 8 {
			r := NewReader(bytes.NewBuffer([]byte{0xff}))
			r.ReadBits(byte(i))
			if err := r.Align(); err != nil {
				t.Fatalf("unexpected error while aligning after reading %d bits: %v", i, err)
			}
			if r.cache != 0 || r.cacheSize != 0 {
				t.Fatalf("cache and cache size should be empty after aligning, cache: %d, cache size: %d", r.cache, r.cacheSize)
			}
		}
	})
}

func TestCombinations(t *testing.T) {
	t.Run("byte reading after reading bits", func(t *testing.T) {
		r := NewReader(bytes.NewBuffer([]byte{0b01000011, 0b10100011}))

		r.ReadBit()
		r.ReadBit()
		r.ReadBit()

		res, err := r.ReadByte()
		if err != nil {
			t.Fatalf("error while reading byte: %v", err)
		}
		if res != 0b00011101 {
			t.Fatalf("invalid byte readed, expected: %#b, got: %#b", 0b00011101, res)
		}
		for range 3 {
			if bit, err := r.ReadBit(); err != nil || bit != 0 {
				t.Fatalf("error while reading bit: %#b, %v", bit, err)
			}
		}
		for range 2 {
			if bit, err := r.ReadBit(); err != nil || bit != 1 {
				t.Fatalf("error while reading bit: %#b, %v", bit, err)
			}
		}
		if _, err := r.ReadBit(); err != io.EOF {
			t.Fatalf("eof expected after reding all bits, got: %v", err)
		}
	})

	t.Run("bytes reading after reading bits", func(t *testing.T) {
		source := []byte{0b01000011, 0b10100011, 0b00110110, 0b01011111}
		r := NewReader(bytes.NewBuffer(source))

		r.ReadBit()
		r.ReadBit()
		r.ReadBit()
		r.ReadBit()

		buffer := make([]byte, 3)
		if _, err := r.Read(buffer); err != nil {
			t.Fatalf("unexpected error whle reading bytes: %v", err)
		}

		expected := []byte{0b00111010, 0b00110011, 0b01100101}
		if !bytes.Equal(expected, buffer) {
			t.Fatalf("invalid bytes readed, expected: %v, got: %v", expected, buffer)
		}

		if r.cache != 0b11110000 || r.cacheSize != 4 {
			t.Fatalf("invalud cache or cache size values, cache: %#b, %d", r.cache, r.cacheSize)
		}

		for range 4 {
			if bit, err := r.ReadBit(); err != nil || bit != 1 {
				t.Fatalf("error while reading bit: %#b, %v", bit, err)
			}
		}

		if _, err := r.ReadBit(); err != io.EOF {
			t.Fatalf("eof expected after reding all bits, got: %v", err)
		}
	})

	t.Run("ReadBits + ReadBit", func(t *testing.T) {
		// brute force
		for k := range 255 {
			expected := byte(k)
			for i := range 8 {
				r := NewReader(bytes.NewBuffer([]byte{expected}))
				mask := byte(((1 << i) - 1) << (8 - i))
				if bits, _ := r.ReadBits(byte(i)); bits != expected&mask {
					t.Fatalf("invalid bits, expected: %#08b, got: %#08b", expected&mask, bits)
				}
				mask = (^mask) << i
				var result byte = 0
				for j := i + 1; j <= 8; j++ {
					bit, err := r.ReadBit()
					if err != nil {
						t.Fatalf("unexpected error occured while reading bit: %v", err)
					}
					result |= bit << (8 - j + i)
				}
				if result != (expected<<i)&mask {
					t.Fatalf("invalid bits, expected: %#08b, got: %#08b", (expected<<i)&mask, result)
				}
			}
		}
	})
}
