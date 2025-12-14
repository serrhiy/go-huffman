package bitio

import (
	"bytes"
	"errors"
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
		if w.cacheSize != 1 {
			t.Fatalf("invalid cache sizw after writing 1 byte, expected 1, got: %d", w.cacheSize)
		}
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

	t.Run("invalid bytes length", func(t *testing.T) {
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

	t.Run("several bytes", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)

		w.WriteBits(0b10110000, 5)
		w.WriteBits(0b01011100, 7)
		if w.cacheSize != 4 {
			t.Fatalf("invalid cache size, expected: %d", w.cacheSize)
		}
		if err := w.Flush(); err != nil {
			t.Fatalf("unexpected error while flushing: %v", err)
		}

		result := buf.Bytes()
		if len(result) != 2 {
			t.Fatalf("invalid buffers length, expected: %d, got: %d", 2, len(result))
		}
		if result[0] != 0b10110010 {
			t.Fatalf("invalid first byte writed, expected: %#08b, got: %#08b", 0b10110101, result[0])
		}
		if result[1] != 0b11100000 {
			t.Fatalf("invalid second byte writed, expected: %#08b, got: %#08b", 0b11101000, result[1])
		}
	})
}

func TestWriteBit(t *testing.T) {
	t.Run("write 1 bit(bit = 1)", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		w.WriteBit(1)
		w.Flush()
		if buf.Bytes()[0] != (1 << 7) {
			t.Fatalf("invalid bit writed, expected: 1, got: 0")
		}
	})

	t.Run("write 1 bit(bit = 0)", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		w.WriteBit(0)
		w.Flush()
		if buf.Bytes()[0] != 0 {
			t.Fatalf("invalid bit writed, expected: 0, got: 1")
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		w := NewWriter(&errWriter{1})
		for range 8 {
			w.WriteBit(1)
		}
		if err := w.Flush(); err != nil {
			t.Fatalf("unexpexted error while flushing: %v", err)
		}
		w.WriteBit(1)
		if err := w.Flush(); err == nil {
			t.Fatal("expected error but got <nil>")
		}
	})

	t.Run("write whole byte", func(t *testing.T) {
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
					t.Fatalf("unexpected error while writing: %v", err)
				}
			}

			if err := w.Flush(); err != nil {
				t.Fatalf("unexpected error while flushing: %v", err)
			}

			if buf.Bytes()[0] != tc.want {
				t.Fatalf("expected %#08b got %#08b", tc.want, buf.Bytes()[0])
			}
		}
	})

	t.Run("several bytes", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)

		w.WriteBit(1)
		w.WriteBit(1)
		w.WriteBit(0)
		w.WriteBit(0)
		w.WriteBit(1)
		w.WriteBit(0)
		w.WriteBit(1)
		w.WriteBit(0)

		w.WriteBit(1)
		w.WriteBit(1)
		w.WriteBit(0)

		if err := w.Flush(); err != nil {
			t.Fatalf("unexpected error while flushing: %v", err)
		}

		result := buf.Bytes()
		if len(result) != 2 {
			t.Fatalf("invalid buffers length, expected: %d, got: %d", 2, len(result))
		}
		if result[0] != 0b11001010 {
			t.Fatalf("invalid first byte writed, expected: %#08b, got: %#08b", 0b11001010, result[0])
		}
		if result[1] != 0b11000000 {
			t.Fatalf("invalid second byte writed, expected: %#08b, got: %#08b", 0b11000000, result[1])
		}
	})
}

func TestFlush(t *testing.T) {
	t.Run("empty flush", func(t *testing.T) {
		w := NewWriter(nil)
		if err := w.Flush(); err != nil {
			t.Fatalf("unexpected error while flusing empty buffer: %v", err)
		}
	})

	// Other cases are tested in TestWrite*
}

func TestWriterAlign(t *testing.T) {
	t.Run("empty align", func(t *testing.T) {
		w := NewWriter(nil)
		if err := w.Align(); err != nil {
			t.Fatalf("unexpected error while aligning empty buffer: %v", err)
		}
	})

	t.Run("align aligned byte", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
		w.WriteByte(0xff)
		if err := w.Align(); err != nil {
			t.Fatalf("unexpected error while aligning: %v", err)
		}
	})

	t.Run("standart case", func(t *testing.T) {
		const expected = 0b11100000
		buf := &bytes.Buffer{}
		w := NewWriter(buf)

		w.WriteBits(expected, 3)

		if err := w.Align(); err != nil {
			t.Fatalf("unexpected error while aligning: %v", err)
		}
		if err := w.Flush(); err != nil {
			t.Fatalf("unexpected wrror while flushing: %v", err)
		}
		if len(buf.Bytes()) != 1 {
			t.Fatalf("expected 1 byte after align, got %d", len(buf.Bytes()))
		}

		got := buf.Bytes()[0]
		if got != expected {
			t.Fatalf("expected %#08b got %#08b", expected, got)
		}
		if w.cacheSize != 0 {
			t.Fatalf("expected cacheSize=0 got %d", w.cacheSize)
		}
	})

	t.Run("variable cache sizes", func(t *testing.T) {
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
	})
}

func TestWriterCombinations(t *testing.T) {
	t.Run("WriteBit and WriteByte", func(t *testing.T) {
		t.Run("WriteBit + WriteByte", func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := NewWriter(buf)

			w.WriteBit(1)
			w.WriteBit(0)
			w.WriteBit(1)

			w.WriteByte(0b10011010)

			w.Flush()

			result := buf.Bytes()
			if len(result) != 2 {
				t.Fatalf("invalid length, expected: %d, got: %d", 2, len(result))
			}
			if result[0]&0b11100000 != 0b10100000 {
				t.Fatalf("invalid first 3 bits writed, expected: %#08b, got: %#08b", 0b10100000, result[0]&0b11100000)
			}
			if result[0] != 0b10110011 {
				t.Fatalf("invalid first byte writed, expected: %#08b, got: %#08b", 0b10110011, result[0])
			}
			if result[1] != 0b01000000 {
				t.Fatalf("invalid second byte writed, expected: %#08b, got: %#08b", 0b01000000, result[1])
			}
		})

		t.Run("WriteByte + WriteBit", func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := NewWriter(buf)

			w.WriteByte(0b10011010)

			w.WriteBit(1)
			w.WriteBit(1)
			w.WriteBit(0)
			w.WriteBit(1)

			w.Flush()

			result := buf.Bytes()
			if len(result) != 2 {
				t.Fatalf("invalid buffer length, expected: %d, got: %d", 2, len(result))
			}
			if result[0] != 0b10011010 {
				t.Fatalf("invalid first byte writed, expected: %#08b, got: %#08b", 0b10011010, result[0])
			}
			if result[1]&0b11110000 != 0b11010000 {
				t.Fatalf("invalid second byte writed, expected: %#08b, got: %#08b", 0b11010000, result[1])
			}
		})
	})

	t.Run("WriteBits and WriteBit", func(t *testing.T) {
		t.Run("inside byte", func(t *testing.T) {
			for i := range 256 {
				expected := byte(i)
				for j := range 8 {
					buf := &bytes.Buffer{}
					w := NewWriter(buf)
					if err := w.WriteBits(expected, byte(j)); err != nil {
						t.Fatalf("unexpected error while writing bits: %v", err)
					}
					for k := j; k < 8; k++ {
						bit := expected & (1 << (8 - 1 - k))
						if err := w.WriteBit(bit); err != nil {
							t.Fatalf("unexpected error while writing bit: %v", err)
						}
					}
					if err := w.Flush(); err != nil {
						t.Fatalf("unexpected error whiule flushing: %v", err)
					}
					result := buf.Bytes()
					if len(result) != 1 {
						t.Fatalf("unexpected buffers length, expected: %d, got: %d", 1, len(result))
					}
					if result[0] != expected {
						t.Fatalf("invalid bytes writed, expected: %#08b, got: %#08b", expected, result[0])
					}
				}
			}
		})

		t.Run("WriteBit + WriteBits", func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := NewWriter(buf)

			w.WriteBit(1)
			w.WriteBit(0)
			w.WriteBit(1)

			w.WriteBits(0b01101100, 6)

			w.Flush()

			result := buf.Bytes()
			if len(result) != 2 {
				t.Fatalf("invalid buffer length, expected: %d, got: %d", 2, len(result))
			}
			if result[0] != 0b10101101 {
				t.Fatalf("invalid first byte writed, expected: %#08b, got: %#08b", 0b10101101, result[0])
			}
			if result[1] != 0b10000000 {
				t.Fatalf("invalid second byte writed, expected: %#08b, got: %#08b", 0b10000000, result[1])
			}
		})

		t.Run("WriteBits + WriteBit", func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := NewWriter(buf)

			w.WriteBits(0b10110000, 5)

			w.WriteBit(1)
			w.WriteBit(0)
			w.WriteBit(0)
			w.WriteBit(1)
			w.WriteBit(0)
			w.WriteBit(1)
			w.WriteBit(1)

			w.Flush()

			result := buf.Bytes()
			if len(result) != 2 {
				t.Fatalf("invalid bytes number writed, expected: %d, got: %d", 2, len(result))
			}
			if result[0] != 0b10110100 {
				t.Fatalf("invalid first byte writed, expected: %#08b, got: %#08b", 0b10110100, result[0])
			}
			if result[1] != 0b10110000 {
				t.Fatalf("invalid second byte writed, expected: %#08b, got: %#08b", 0b10110000, result[1])
			}
		})
	})

	t.Run("WriteBits + Write", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)

		w.WriteBit(1)
		w.WriteBit(0)
		w.WriteBit(0)
		w.WriteBit(1)

		w.Write([]byte{0xff})

		w.Flush()

		result := buf.Bytes()
		if len(result) != 2 {
			t.Fatalf("invalid bytes number writed, expected: %d, got: %d", 2, len(result))
		}
		if result[0] != 0b10011111 {
			t.Fatalf("invalid first byte writed, expected: %#08b, got: %#08b", 0b10011111, result[0])
		}
		if result[1] != 0b11110000 {
			t.Fatalf("invalid second byte writed, expected: %#08b, got: %#08b", 0b11110000, result[1])
		}
	})
}
