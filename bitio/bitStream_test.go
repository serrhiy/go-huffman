package bitio

import (
	"slices"
	"testing"
)

func TestBitStream(t *testing.T) {
	t.Run("BitStream.Push", func(t *testing.T) {
		t.Run("1 bit(1)", func(t *testing.T) {
			stream := NewBitStream()
			stream.Push(1)
			if stream.Len() != 1 {
				t.Fatalf("unexpected length, expected: %d, got: %d", 1, stream.Len())
			}
			if stream.position != 1 {
				t.Fatalf("invalid stream position, expected: %d, got: %d", 1, stream.position)
			}
			if len(stream.bits) != 1 || stream.bits[0] != uint64(1)<<63 {
				t.Fatalf("invalid stream.bits valud, expected: [0b10000000], got: %v", stream.bits)
			}
		})

		t.Run("1 bit(0)", func(t *testing.T) {
			stream := NewBitStream()
			stream.Push(0)
			if stream.Len() != 1 {
				t.Fatalf("unexpected length, expected: %d, got: %d", 1, stream.Len())
			}
			if stream.position != 1 {
				t.Fatalf("invalid stream position, expected: %d, got: %d", 1, stream.position)
			}
			if len(stream.bits) != 1 || stream.bits[0] != 0 {
				t.Fatalf("invalid stream.bits valud, expected: [0], got: %v", stream.bits)
			}
		})

		t.Run("inside uint64", func(t *testing.T) {
			testCases := []struct {
				name   string
				number uint64
			}{
				{"all_zeroes", 0},
				{"all_ones", 0xffffffffffffffff},
				{"upper_half_ones", 0xffffffff00000000},
				{"lower_half_ones", 0x00000000ffffffff},
				{"alternating_blocks", 0xf00f00f000f00f0f},
			}
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					stream := NewBitStream()
					for i := range 64 {
						bit := tc.number >> (63 - i) & 1
						stream.Push(byte(bit))
					}
					if stream.Len() != 64 {
						t.Fatalf("invalid length, expected: %d, got: %d", 64, stream.Len())
					}
					if stream.bits[0] != tc.number {
						t.Fatalf("invalid bits pushed, expected: %d, got: %d", tc.number, stream.bits[0])
					}
				})
			}
		})

		t.Run("several uint64", func(t *testing.T) {
			testCases := []struct {
				name    string
				numbers []uint64
			}{
				{"all_zeroes", []uint64{0, 0, 0}},
				{"all_ones", []uint64{0xffffffffffffffff, 0xffffffffffffffff}},
				{"upper_half_ones", []uint64{0xffffffff00000000, 0x00000000ffffffff}},
				{"lower_half_ones", []uint64{0x00000000ffffffff, 0xffffffff00000000}},
				{"alternating_blocks", []uint64{0xf00f0f0f0f00f0ff, 0xff0f0ff0f00f0f0f}},
			}
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					stream := NewBitStream()
					for _, number := range tc.numbers {
						for i := range 64 {
							bit := number >> (63 - i) & 1
							stream.Push(byte(bit))
						}
					}
					if stream.Len() != 64*len(tc.numbers) {
						t.Fatalf("invalid length, expected: %d, got: %d", 64*len(tc.numbers), stream.Len())
					}
					for i := range tc.numbers {
						if tc.numbers[i] != stream.bits[i] {
							t.Fatalf("invalid bits pushed, expected: %v, got: %v", tc.numbers, stream.bits)
						}
					}
				})
			}
		})

		t.Run("across uint64", func(t *testing.T) {
			stream := NewBitStream()
			var number uint64 = 0xff0f0ff0f00f0f0f
			for i := range 64 {
				bit := number >> (63 - i) & 1
				stream.Push(byte(bit))
			}
			if stream.Len() != 64 {
				t.Fatalf("wrong len, expected: %d, got: %d", 64, stream.Len())
			}
			if stream.bits[0] != number {
				t.Fatalf("invalid bits pushed, expected: %d, got: %d", number, stream.bits[0])
			}

			stream.Push(1)
			if stream.Len() != 65 {
				t.Fatalf("wrong len, expected: %d, got: %d", 65, stream.Len())
			}
			if stream.bits[1] != uint64(1)<<63 {
				t.Fatalf("invalid bits pushed, expected: %d, got: %d", uint64(1)<<63, stream.bits[1])
			}
		})
	})

	t.Run("Len", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			stream := NewBitStream()
			if stream.Len() != 0 {
				t.Fatalf("it is expected that the stream will be empty, got: %d", stream.Len())
			}
		})

		t.Run("1 bit", func(t *testing.T) {
			stream := NewBitStream()
			stream.Push(1)
			if stream.Len() != 1 {
				t.Fatalf("wrong length, expected: %d, got: %d", 1, stream.Len())
			}
		})
		// other cases are testes in Push
	})

	t.Run("Iterator", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			stream := NewBitStream()
			for range stream.Iter() {
				t.Fatal("iterator must be empty on an empty stream")
			}
		})

		t.Run("1 bit", func(t *testing.T) {
			stream := NewBitStream()
			stream.Push(1)
			all := slices.Collect(stream.Iter())
			if len(all) != 1 || all[0] != 1 {
				t.Fatalf("invalid iterator values, expected: [1], got: %v", all)
			}
		})

		t.Run("uint64", func(t *testing.T) {
			stream := NewBitStream()
			var number uint64 = 0xff0f0ff0f00f0f0f
			for i := range 64 {
				bit := number >> (63 - i) & 1
				stream.Push(byte(bit))
			}
			var result uint64 = 0
			var index uint64 = 0
			for bit := range stream.Iter() {
				result |= uint64(bit) << (63 - index)
				index++
			}
			if result != number {
				t.Fatalf("invalid number readed, expected: %d, got: %d", number, result)
			}
		})

		t.Run("several uint64", func(t *testing.T) {
			numbers := []uint64{0xf00f0f0f0f00f0ff, 0xff0f0ff0f00f0f0f}
			stream := NewBitStream()
			for _, number := range numbers {
				for i := range 64 {
					bit := number >> (63 - i) & 1
					stream.Push(byte(bit))
				}
			}
			result := make([]uint64, 0, len(numbers))
			var number uint64 = 0
			var index uint64 = 0
			for bit := range stream.Iter() {
				number |= uint64(bit) << (63 - index)
				index++
				if index == 64 {
					result = append(result, number)
					number = 0
					index = 0
				}
			}

			for index := range numbers {
				if numbers[index] != result[index] {
					t.Fatalf("invalid bits readed, expected: %v, got: %v", numbers, result)
				}
			}
		})
	})
}
