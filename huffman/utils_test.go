package huffman

import (
	"bytes"
	"container/heap"
	"testing"
)

func TestToPriorityQueue(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		pq := toPriorityQueue(map[byte]uint{})

		if pq.Len() != 0 {
			t.Fatalf("expected empty priority queue, got %d", pq.Len())
		}
	})

	t.Run("single element", func(t *testing.T) {
		freq := map[byte]uint{'a': 5}
		pq := toPriorityQueue(freq)

		if pq.Len() != 1 {
			t.Fatalf("expected queue length 1, got %d", pq.Len())
		}

		item, ok := heap.Pop(&pq).(*node)
		if !ok || item.char != 'a' || item.count != 5 {
			t.Fatalf("unexpected element: got {%c, %d}", item.char, item.count)
		}
	})

	t.Run("multiple items", func(t *testing.T) {
		freq := map[byte]uint{
			'a': 5,
			'b': 2,
			'c': 9,
			'd': 1,
		}

		pq := toPriorityQueue(freq)

		if pq.Len() != 4 {
			t.Fatalf("expected length 4, got %d", pq.Len())
		}

		expectedOrder := []struct {
			char  byte
			count uint
		}{
			{'d', 1},
			{'b', 2},
			{'a', 5},
			{'c', 9},
		}

		for i, exp := range expectedOrder {
			item, ok := heap.Pop(&pq).(*node)
			if !ok || item.char != exp.char || item.count != exp.count {
				t.Fatalf("at pop %d: expected {%c, %d}, got {%c, %d}",
					i, exp.char, exp.count, item.char, item.count)
			}
		}
	})
}

func TestBuildTree(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if root := buildTree(map[byte]uint{}); root != nil {
			t.Fatalf("non <nil> result on empty frequency map: %v", root)
		}
	})

	t.Run("single node", func(t *testing.T) {
		root := buildTree(map[byte]uint{'a': 5})

		if root == nil {
			t.Fatal("expected non-nil root")
		}
		if root.isLeaf() {
			t.Fatalf("expected internal node, got leaf node")
		}
		if root.left.char != 'a' || root.left.count != 5 {
			t.Fatalf("unexpected values: got (%c,%d)", root.char, root.count)
		}
	})

	t.Run("two nodes", func(t *testing.T) {
		root := buildTree(map[byte]uint{'a': 2, 'b': 3})
		if root == nil {
			t.Fatal("tree root is nil")
		}
		if root.count != 5 {
			t.Fatalf("invalid root count: got %d, expected %d", root.count, 5)
		}

		if !root.left.isLeaf() || root.left.char != 'a' {
			t.Fatal("invalid tree structure, root.left expected to be 'a'")
		}
		if !root.right.isLeaf() || root.right.char != 'b' {
			t.Fatal("invalid tree structure, root.right expected to be 'b'")
		}
	})

	t.Run("tree", func(t *testing.T) {
		//        (*,37)
		//      /         \
		// D(15)          (*,22)
		//                /     \
		//           C(10)      (*,12)
		//                      /    \
		//                  A(5)      B(7)
		freq := map[byte]uint{
			'a': 5,
			'b': 7,
			'c': 10,
			'd': 15,
		}

		root := buildTree(freq)

		if root == nil {
			t.Fatal("tree root is nil")
		}

		expected := uint(5 + 7 + 10 + 15)
		if root.count != expected {
			t.Fatalf("invalid root count: got %d, expected %d", root.count, expected)
		}
		if !root.left.isLeaf() || root.left.char != 'd' {
			t.Fatal("invalid tree structure, root.left expected to be 'd'")
		}
		if root.right.isLeaf() {
			t.Fatal("invalid tree structure, root.right expected to be internal")
		}
		if !root.right.left.isLeaf() || root.right.left.char != 'c' {
			t.Fatal("invalid tree structure, root.right.left expected to be 'c'")
		}
		if root.right.right.isLeaf() {
			t.Fatal("invalid tree structure, root.right.right expected to be internal")
		}
		if !root.right.right.left.isLeaf() || root.right.right.left.char != 'a' {
			t.Fatal("invalid tree structure, root.right.right.left expected to be 'a'")
		}
		if !root.right.right.right.isLeaf() || root.right.right.right.char != 'b' {
			t.Fatal("invalid tree structure, root.right.right.right expected to be 'b'")
		}
	})
}

func TestBuildCodes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if codes := buildCodes(nil); len(codes) != 0 {
			t.Fatal("empty codes expected on <nil> root")
		}
	})

	// this case is unreachable when the input originates from buildTree
	t.Run("single leaf and no internal node", func(t *testing.T) {
		root := &node{char: 'a', count: 10}
		codes := buildCodes(root)
		if len(codes) != 1 {
			t.Fatalf("expected 1 code, got %d", len(codes))
		}

		if code, ok := codes['a']; !ok || code != "" {
			t.Fatalf("single leaf must have empty code, got %q, %v", code, codes)
		}
	})

	t.Run("two leafes", func(t *testing.T) {
		root := &node{
			left:  &node{char: 'a', count: 10},
			right: &node{char: 'b', count: 15},
		}
		codes := buildCodes(root)

		if len(codes) != 2 {
			t.Fatalf("invalid codes map size, expected: %d, got: %d", 1, len(codes))
		}
		if code, ok := codes['a']; !ok || code != "1" {
			t.Fatalf("invalide code map builded, expected: %s, got: %s", "1", code)
		}
		if code, ok := codes['b']; !ok || code != "0" {
			t.Fatalf("invalide code map builded, expected: %s, got: %s", "0", code)
		}
	})

	t.Run("tree", func(t *testing.T) {
		//        (*,37)
		//    1 /         \ 0
		// D(15)          (*,22)
		//              1 /     \ 0
		//           C(10)      (*,12)
		//                    1 /    \ 0
		//                  A(5)      B(7)
		root := &node{
			left: &node{char: 'd', count: 15},
			right: &node{
				count: 22,
				left:  &node{char: 'c', count: 10},
				right: &node{
					count: 12,
					left:  &node{char: 'a', count: 5},
					right: &node{char: 'b', count: 7},
				},
			},
		}

		codes := buildCodes(root)
		if len(codes) != 4 {
			t.Fatalf("invalid codes map size, expected: %d, got: %d", 4, len(codes))
		}

		expected := map[byte]string{
			'd': "1",
			'c': "01",
			'a': "001",
			'b': "000",
		}

		for char, code := range expected {
			if actual, ok := codes[char]; !ok || actual != code {
				t.Fatalf("invalud code builded, expected: %s, actual: %s", code, actual)
			}
		}
	})
}

func TestBuildReverseCodesSingleLeaf(t *testing.T) {
	root := &node{char: 'A'}
	codes := buildReverseCodes(root)

	if len(codes) != 1 {
		t.Fatalf("expected 1 code, got %d", len(codes))
	}
	if codes[""] != 'A' {
		t.Fatalf("expected code '' -> 'A', got %v", codes)
	}
}

func TestBuildReverseCodesTwoLeaves(t *testing.T) {
	root := &node{
		left:  &node{char: 'A'},
		right: &node{char: 'B'},
	}
	codes := buildReverseCodes(root)

	if len(codes) != 2 {
		t.Fatalf("expected 2 codes, got %d", len(codes))
	}
	if codes["1"] != 'A' || codes["0"] != 'B' {
		t.Fatalf("incorrect codes: %v", codes)
	}
}

func TestBuildReverseCodesThreeLeaves(t *testing.T) {
	root := &node{
		left: &node{
			left:  &node{char: 'A'},
			right: &node{char: 'B'},
		},
		right: &node{char: 'C'},
	}
	codes := buildReverseCodes(root)

	expected := map[string]byte{
		"11": 'A',
		"10": 'B',
		"0":  'C',
	}

	if len(codes) != 3 {
		t.Fatalf("expected 3 codes, got %d", len(codes))
	}

	for k, v := range expected {
		if codes[k] != v {
			t.Fatalf("expected code %s -> %c, got %v", k, v, codes)
		}
	}
}

func TestBuildReverseCodesNilTree(t *testing.T) {
	codes := buildReverseCodes(nil)
	if len(codes) != 0 {
		t.Fatalf("expected empty map for nil tree, got %v", codes)
	}
}

func TestBuildReverseCodesComplexTree(t *testing.T) {
	root := &node{
		left: &node{
			left:  &node{char: 'A'},
			right: &node{char: 'B'},
		},
		right: &node{
			left:  &node{char: 'C'},
			right: &node{char: 'D'},
		},
	}
	codes := buildReverseCodes(root)

	expected := map[string]byte{
		"11": 'A',
		"10": 'B',
		"01": 'C',
		"00": 'D',
	}

	if len(codes) != 4 {
		t.Fatalf("expected 4 codes, got %d", len(codes))
	}

	for k, v := range expected {
		if codes[k] != v {
			t.Fatalf("expected code %s -> %c, got %v", k, v, codes)
		}
	}
}

func TestCalculateContentSizeEmpty(t *testing.T) {
	codes := make(map[byte]string)
	frequencies := make(map[byte]uint)
	size, err := calculateContentSize(codes, frequencies)
	if err != nil {
		t.Fatalf("unxepected error: %v", err)
	}
	if size != 0 {
		t.Fatalf("empty calculateContentSize failed, expected 0, got %d", size)
	}
}

func TestCalculateContentSize(t *testing.T) {
	codes := map[byte]string{
		'a': "1",
		'b': "01",
		'c': "00",
	}
	frequencies := map[byte]uint{
		'a': 10,
		'b': 5,
		'c': 4,
	}
	const expected = 10 + 5*2 + 4*2
	size, err := calculateContentSize(codes, frequencies)
	if err != nil {
		t.Fatalf("unxepected error: %v", err)
	}
	if size != expected {
		t.Fatalf("TestCalculateContentSize failed, expected: %d, got %d", expected, size)
	}
}

func TestCalculateContentSizeReal(t *testing.T) {
	source := []byte("aaaaaaaaaabbbbbcccc")
	reader := bytes.NewReader(source)
	encoder := NewEncoder(reader, nil)
	frequencies, err := encoder.getFrequencyMap()
	if err != nil {
		t.Fatalf("unexpected getFrequencyMap error: %v", err)
	}
	codes := buildCodes(buildTree(frequencies))
	const expected = 10 + 5*2 + 4*2
	size, err := calculateContentSize(codes, frequencies)
	if err != nil {
		t.Fatalf("unxepected error: %v", err)
	}
	if size != expected {
		t.Fatalf("TestCalculateContentSizeReal failed, expected: %d, got %d", expected, size)
	}
}

func TestCalculateContentSizeError(t *testing.T) {
	codes := map[byte]string{
		'a': "1",
		'b': "01",
		'c': "00",
	}
	frequencies := map[byte]uint{
		'a': 10,
		'b': 5,
		// 'c' is absent
	}
	if size, err := calculateContentSize(codes, frequencies); err == nil {
		t.Fatalf("error expected but got: %d, %v", size, err)
	}
}

func TestCalculateTreeSizeEmpty(t *testing.T) {
	if size := calculateTreeSize(nil); size != 0 {
		t.Fatalf("expected 0 but got %d", size)
	}
}

func TestCalculateTreeSize(t *testing.T) {
	const expected = 1 + 9 + 1 + 9 + 9
	root := &node{
		left: &node{
			char:  'a',
			count: 10,
		},
		right: &node{
			left: &node{
				char:  'b',
				count: 5,
			},
			right: &node{
				char:  'c',
				count: 4,
			},
		},
	}
	if size := calculateTreeSize(root); size != expected {
		t.Fatalf("calculateTreeSize faled, expected: %d, got %d", expected, size)
	}
}
