package huffman

import (
	"container/heap"
	"testing"
)

func TestToPriorityQueueEmpty(t *testing.T) {
	freq := map[byte]uint{}
	pq := toPriorityQueue(freq)

	if pq.Len() != 0 {
		t.Fatalf("expected empty priority queue, got %d", pq.Len())
	}
}

func TestToPriorityQueueSingleElement(t *testing.T) {
	freq := map[byte]uint{'A': 5}
	pq := toPriorityQueue(freq)

	if pq.Len() != 1 {
		t.Fatalf("expected queue length 1, got %d", pq.Len())
	}

	item, ok := heap.Pop(&pq).(*node)
	if !ok || item.char != 'A' || item.count != 5 {
		t.Fatalf("unexpected element: got {%c, %d}", item.char, item.count)
	}
}

func TestToPriorityQueueMultiple(t *testing.T) {
	freq := map[byte]uint{
		'A': 5,
		'B': 2,
		'C': 9,
		'D': 1,
	}

	pq := toPriorityQueue(freq)

	if pq.Len() != 4 {
		t.Fatalf("expected length 4, got %d", pq.Len())
	}

	expectedOrder := []struct {
		char  byte
		count uint
	}{
		{'D', 1},
		{'B', 2},
		{'A', 5},
		{'C', 9},
	}

	for i, exp := range expectedOrder {
		item, ok := heap.Pop(&pq).(*node)
		if !ok || item.char != exp.char || item.count != exp.count {
			t.Fatalf("at pop %d: expected {%c, %d}, got {%c, %d}",
				i, exp.char, exp.count, item.char, item.count)
		}
	}
}
