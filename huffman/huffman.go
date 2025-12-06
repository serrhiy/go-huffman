package huffman

import (
	"bufio"
	"container/heap"
	"io"
	"os"
)

const bufferSize = 32 * 1024

type node struct {
	char  byte
	count uint64
	left  *node
	right *node
}

type Huffman struct {
	probabilities map[byte]uint64
}

func buildCodes(root *node, prefix string, table map[byte]string) {
	if root == nil {
		return
	}
	if root.left == nil && root.right == nil {
		table[root.char] = prefix
		return
	}
	buildCodes(root.left, prefix+"1", table)
	buildCodes(root.right, prefix+"0", table)
}

func NewFromFile(filepath string) (*Huffman, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := make(map[byte]uint64)
	reader := bufio.NewReader(file)
	buffer := make([]byte, bufferSize)
	for {
		bytes, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		for i := range bytes {
			result[buffer[i]]++
		}
	}
	return &Huffman{result}, nil
}

func (huffman Huffman) Encode() {
	root := huffman.buildTree()
	table := make(map[byte]string)
	buildCodes(root, "", table)
}

func (huffman Huffman) buildTree() *node {
	queue := toPriorityQueue(huffman.probabilities)
	for queue.Len() > 1 {
		node1 := heap.Pop(&queue).(*node)
		node2 := heap.Pop(&queue).(*node)
		combined := &node{0, node1.count + node2.count, node1, node2}
		heap.Push(&queue, combined)
	}
	return heap.Pop(&queue).(*node)
}
