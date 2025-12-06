package huffman

import (
	"container/heap"
)

func toPriorityQueue(probability map[byte]uint64) priorityQueue {
	result := make(priorityQueue, 0, len(probability))
	for char, count := range probability {
		result = append(result, &node{char, count, nil, nil})
	}
	heap.Init(&result)
	return result
}
