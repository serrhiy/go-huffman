package benchkit

import (
	"math/rand/v2"
)

func Text(size int) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = byte('a' + i%26)
	}
	return string(b)
}

func Random(size int) string {
	seed := [32]byte{31: 207}
	chacha := rand.NewChaCha8(seed)
	b := make([]byte, size)
	chacha.Read(b)
	return string(b)
}

func Range(start, end int) string {
	if end <= start {
		return ""
	}
	b := make([]byte, end-start)
	for i := range b {
		b[i] = byte(start + i)
	}
	return string(b)
}
