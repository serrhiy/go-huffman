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

func sumLeafCounts(n *node) uint {
	if n == nil {
		return 0
	}
	if n.isLeaf() {
		return n.count
	}
	return sumLeafCounts(n.left) + sumLeafCounts(n.right)
}

func validateInternalCounts(t *testing.T, n *node) uint {
	if n.isLeaf() {
		return n.count
	}
	left := validateInternalCounts(t, n.left)
	right := validateInternalCounts(t, n.right)

	if n.count != left+right {
		t.Fatalf("invalid internal node count: got %d, expected %d", n.count, left+right)
	}
	return n.count
}

func TestBuildTreeSingleNode(t *testing.T) {
	freq := map[byte]uint{
		'a': 5,
	}

	root := buildTree(freq)

	if root == nil {
		t.Fatal("expected non-nil root")
	}
	if !root.isLeaf() {
		t.Fatalf("expected leaf node, got internal node")
	}
	if root.char != 'a' || root.count != 5 {
		t.Fatalf("unexpected values: got (%c,%d)", root.char, root.count)
	}
}

func TestBuildTreeBasic(t *testing.T) {
	freq := map[byte]uint{
		'a': 5,
		'b': 7,
		'c': 10,
		'd': 15,
	}

	root := buildTree(freq)

	if root == nil {
		t.Fatalf("tree root is nil")
	}

	expected := uint(5 + 7 + 10 + 15)
	if root.count != expected {
		t.Fatalf("invalid root count: got %d, expected %d", root.count, expected)
	}

	validateInternalCounts(t, root)
}

func TestBuildTreeDeterministic(t *testing.T) {
	freq1 := map[byte]uint{
		'a': 5, 'b': 9, 'c': 12, 'd': 13,
	}

	freq2 := map[byte]uint{
		'd': 13, 'c': 12, 'b': 9, 'a': 5,
	}

	tree1 := buildTree(freq1)
	tree2 := buildTree(freq2)

	sum1 := sumLeafCounts(tree1)
	sum2 := sumLeafCounts(tree2)

	if sum1 != sum2 {
		t.Fatalf("trees differ: leaf sums %d vs %d", sum1, sum2)
	}
}
