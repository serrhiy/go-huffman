package huffman

import (
	"container/heap"
	"testing"
)

func newNode(char byte, count uint) *node {
	return &node{char: char, count: count}
}

func TestPriorityQueueLen(t *testing.T) {
	pq := priorityQueue{}
	if pq.Len() != 0 {
		t.Fatalf("expected length 0, got %d", pq.Len())
	}

	heap.Push(&pq, newNode('a', 1))
	if pq.Len() != 1 {
		t.Fatalf("expected length 1, got %d", pq.Len())
	}
}

func TestPriorityQueueLess(t *testing.T) {
	pq := priorityQueue{
		newNode('a', 1),
		newNode('b', 2),
		newNode('c', 0),
	}

	if !pq.Less(2, 0) {
		t.Fatalf("expected pq[2] < pq[0]")
	}
	if pq.Less(1, 0) {
		t.Fatalf("expected pq[1] >= pq[0]")
	}
}

func TestPriorityQueueSwap(t *testing.T) {
	pq := priorityQueue{
		newNode('a', 1),
		newNode('b', 2),
	}
	pq.Swap(0, 1)
	if pq[0].char != 'b' || pq[1].char != 'a' {
		t.Fatalf("Swap failed, got %v %v", pq[0].char, pq[1].char)
	}
}

func TestPriorityQueuePushPopSingle(t *testing.T) {
	var pq priorityQueue
	heap.Init(&pq)

	n := newNode('x', 5)
	heap.Push(&pq, n)

	if pq.Len() != 1 {
		t.Fatalf("expected length 1, got %d", pq.Len())
	}

	item := pq.Pop().(*node)
	if item != n {
		t.Fatalf("expected popped node to be the same")
	}

	if pq.Len() != 0 {
		t.Fatalf("expected length 0 after pop, got %d", pq.Len())
	}
}

func TestPriorityQueueHeapOrder(t *testing.T) {
	var pq priorityQueue
	heap.Init(&pq)

	nodes := []*node{
		newNode('a', 5),
		newNode('b', 1),
		newNode('c', 3),
		newNode('d', 2),
	}

	for _, n := range nodes {
		heap.Push(&pq, n)
	}

	expectedOrder := []byte{'b', 'd', 'c', 'a'}
	for _, expChar := range expectedOrder {
		n := heap.Pop(&pq).(*node)
		if n.char != expChar {
			t.Fatalf("expected %c, got %c", expChar, n.char)
		}
	}
}

func TestPriorityQueuePopEmpty(t *testing.T) {
	var pq priorityQueue
	heap.Init(&pq)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when popping from empty heap")
		}
	}()

	heap.Pop(&pq)
}

func TestPriorityQueueLarge(t *testing.T) {
	var pq priorityQueue
	heap.Init(&pq)

	count := 1000
	for i := range count {
		heap.Push(&pq, newNode(byte(i%256), uint(count-i)))
	}

	prevCount := uint(0)
	for pq.Len() > 0 {
		n := heap.Pop(&pq).(*node)
		if n.count < prevCount {
			t.Fatalf("heap property violated: %d < %d", n.count, prevCount)
		}
		prevCount = n.count
	}
}

func TestNodeIsLeaf(t *testing.T) {
	leaf := &node{char: 'x', count: 1}
	if !leaf.isLeaf() {
		t.Fatalf("leaf node detected as non-leaf")
	}

	parent := &node{left: leaf, right: nil}
	if parent.isLeaf() {
		t.Fatalf("parent node incorrectly detected as leaf")
	}

	parent2 := &node{left: leaf, right: &node{}}
	if parent2.isLeaf() {
		t.Fatalf("parent node incorrectly detected as leaf")
	}
}
