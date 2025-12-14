package huffman

import (
	"bytes"
	"container/heap"
	"testing"
)

func leaf(char byte, count uint) *node {
	return &node{char: char, count: count}
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

func TestBuildTreeSingleNode(t *testing.T) {
	freq := map[byte]uint{
		'a': 5,
	}

	root := buildTree(freq)

	if root == nil {
		t.Fatal("expected non-nil root")
	}
	if root.isLeaf() {
		t.Fatalf("expected internal node, got leaf node")
	}
	if root.left.char != 'a' || root.left.count != 5 {
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

func TestBuildCodesSingleLeaf(t *testing.T) {
	root := leaf('x', 5)

	codes := buildCodes(root)

	if len(codes) != 1 {
		t.Fatalf("expected 1 code, got %d", len(codes))
	}

	if code := codes['x']; code != "" {
		t.Fatalf("single leaf must have empty code, got %q", code)
	}
}

func TestBuildCodesTwoLeaves(t *testing.T) {
	root := &node{
		left:  leaf('A', 2),
		right: leaf('B', 3),
	}

	codes := buildCodes(root)

	if codes['A'] != "1" {
		t.Fatalf("expected A = 1, got %q", codes['A'])
	}
	if codes['B'] != "0" {
		t.Fatalf("expected B = 0, got %q", codes['B'])
	}
}

func TestBuildCodesMultiLevel(t *testing.T) {
	//         (*)
	//       /     \
	//     (*)      C
	//    /   \
	//   A     B
	root := &node{
		left: &node{
			left:  leaf('A', 5),
			right: leaf('B', 7),
		},
		right: leaf('C', 10),
	}

	codes := buildCodes(root)

	if codes['A'] != "11" {
		t.Fatalf("expected A -> 11, got %q", codes['A'])
	}
	if codes['B'] != "10" {
		t.Fatalf("expected B -> 10, got %q", codes['B'])
	}
	if codes['C'] != "0" {
		t.Fatalf("expected C -> 0, got %q", codes['C'])
	}
}

func TestBuildCodesUniqueness(t *testing.T) {
	root := &node{
		left: leaf('A', 1),
		right: &node{
			left:  leaf('B', 2),
			right: leaf('C', 3),
		},
	}

	codes := buildCodes(root)

	seen := map[string]bool{}
	for ch, code := range codes {
		if seen[code] {
			t.Fatalf("duplicate code %q for %q", code, ch)
		}
		seen[code] = true
	}
}

func TestBuildCodesNoPrefixes(t *testing.T) {
	root := &node{
		left: &node{
			left:  leaf('A', 1),
			right: leaf('B', 2),
		},
		right: leaf('C', 3),
	}

	codes := buildCodes(root)

	for k1, c1 := range codes {
		for k2, c2 := range codes {
			if k1 == k2 {
				continue
			}
			if len(c1) < len(c2) && c2[:len(c1)] == c1 {
				t.Fatalf("code %q is prefix of %q", c1, c2)
			}
		}
	}
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
