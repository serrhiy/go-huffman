package huffman

type node struct {
	char  byte
	count uint
	left  *node
	right *node
}

func (n node) isLeaf() bool {
	return n.left == nil && n.right == nil
}
