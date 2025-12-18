package bitio

import (
	"iter"
)

type BitStream struct {
	bits     []uint64
	position uint8
}

func NewBitStream() *BitStream {
	return &BitStream{[]uint64{0}, 0}
}

func BitStreamCopy(stream *BitStream) *BitStream {
	bits := make([]uint64, len(stream.bits))
	copy(bits, stream.bits)
	return &BitStream{bits, stream.position}
}

func (stream *BitStream) Len() int {
	return (len(stream.bits)-1)*64 + int(stream.position)
}

func (stream *BitStream) Push(bit byte) {
	bit &= 1

	index := len(stream.bits) - 1
	stream.bits[index] |= uint64(bit) << (63 - uint64(stream.position))
	stream.position += 1
	if stream.position >= 64 {
		stream.bits = append(stream.bits, 0)
		stream.position = 0
	}
}

func (stream *BitStream) Iter() iter.Seq[byte] {
	return func(yield func(byte) bool) {
		for i, bitset := range stream.bits {
			var stop byte = 64
			if i == len(stream.bits)-1 {
				stop = stream.position
			}

			for j := byte(0); j < stop; j++ {
				bit := (bitset >> (63 - j)) & 1
				if !yield(byte(bit)) {
					return
				}
			}
		}
	}
}
