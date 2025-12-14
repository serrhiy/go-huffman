package huffman

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"

	"github.com/serrhiy/go-huffman/bitio"
)

func getFrequencyMap(r io.Reader) (map[byte]uint, error) {
	result := make(map[byte]uint, 1<<7)
	reader := bufio.NewReader(r)
	buffer := make([]byte, bufferSize)
	for {
		readed, err := reader.Read(buffer)
		for i := range readed {
			result[buffer[i]] += 1
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return result, nil
}

func toPriorityQueue(frequencies map[byte]uint) priorityQueue {
	result := make(priorityQueue, 0, len(frequencies))
	for char, count := range frequencies {
		result = append(result, &node{char, count, nil, nil})
	}
	heap.Init(&result)
	return result
}

func buildTree(frequencies map[byte]uint) *node {
	if len(frequencies) == 0 {
		return nil
	}
	queue := toPriorityQueue(frequencies)
	if queue.Len() == 1 {
		left := heap.Pop(&queue).(*node)
		return &node{0, left.count, left, nil}
	}
	for queue.Len() > 1 {
		node1 := heap.Pop(&queue).(*node)
		node2 := heap.Pop(&queue).(*node)
		combined := &node{0, node1.count + node2.count, node1, node2}
		heap.Push(&queue, combined)
	}
	return heap.Pop(&queue).(*node)
}

func calculateTreeSize(root *node) uint16 {
	if root == nil {
		return 0
	}
	if root.isLeaf() {
		return 1 + 8
	}
	return 1 + calculateTreeSize(root.left) + calculateTreeSize(root.right)
}

func calculateContentSize(codes map[byte]string, frequencies map[byte]uint) (uint64, error) {
	var size uint64 = 0
	for char, code := range codes {
		frequency, ok := frequencies[char]
		if !ok {
			return 0, fmt.Errorf("char %q exists in codes bit absent in frequency map", char)
		}
		size += uint64(frequency * uint(len(code)))
	}
	return size, nil
}

func _buildCodes(root *node, prefix string, table map[byte]string) {
	if root == nil {
		return
	}
	if root.left == nil && root.right == nil {
		table[root.char] = prefix
		return
	}
	_buildCodes(root.left, prefix+"1", table)
	_buildCodes(root.right, prefix+"0", table)
}

func buildCodes(root *node) map[byte]string {
	table := make(map[byte]string, 1<<7)
	_buildCodes(root, "", table)
	return table
}

func _buildReverseCodes(root *node, prefix string, table map[string]byte) {
	if root == nil {
		return
	}
	if root.left == nil && root.right == nil {
		table[prefix] = root.char
		return
	}
	_buildReverseCodes(root.left, prefix+"1", table)
	_buildReverseCodes(root.right, prefix+"0", table)
}

func buildReverseCodes(root *node) map[string]byte {
	table := make(map[string]byte, 1<<7)
	_buildReverseCodes(root, "", table)
	return table
}

func writeCodes(root *node, writer *bitio.Writer) error {
	if root == nil {
		return nil
	}
	if root.isLeaf() {
		if err := writer.WriteBit(1); err != nil {
			return err
		}
		if err := writer.WriteByte(root.char); err != nil {
			return err
		}
		return nil
	}
	if err := writer.WriteBit(0); err != nil {
		return err
	}
	if err := writeCodes(root.left, writer); err != nil {
		return err
	}
	if err := writeCodes(root.right, writer); err != nil {
		return err
	}
	return nil
}
