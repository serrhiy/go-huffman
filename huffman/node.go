package huffman

type node struct {
	char  byte
	count uint
	left  *node
	right *node
}
