package huffman

import (
	"container/heap"
)

func toPriorityQueue(frequencies map[byte]uint) priorityQueue {
	result := make(priorityQueue, 0, len(frequencies))
	for char, count := range frequencies {
		result = append(result, &node{char, count, nil, nil})
	}
	heap.Init(&result)
	return result
}

func buildTree(frequencies map[byte]uint) *node {
	queue := toPriorityQueue(frequencies)
	for queue.Len() > 1 {
		node1 := heap.Pop(&queue).(*node)
		node2 := heap.Pop(&queue).(*node)
		combined := &node{0, node1.count + node2.count, node1, node2}
		heap.Push(&queue, combined)
	}
	return heap.Pop(&queue).(*node)
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
	table := make(map[byte]string, 1<<8)
	_buildCodes(root, "", table)
	return table
}
