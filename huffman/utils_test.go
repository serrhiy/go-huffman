package huffman

import (
	"bytes"
	"container/heap"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/serrhiy/go-huffman/benchkit"
	"github.com/serrhiy/go-huffman/bitio"
)

type errorReader struct {
	limit  int
	reader io.Reader
}

func (r *errorReader) Read(p []byte) (int, error) {
	if r.limit <= 0 {
		return 0, errors.New("reading limit reached")
	}
	toRead := min(len(p), r.limit)
	buffer := make([]byte, toRead)
	r.limit -= toRead
	return r.reader.Read(buffer)
}

func TestGetFrequencyMap(t *testing.T) {
	t.Run("empty reader", func(t *testing.T) {
		buf := &bytes.Buffer{}
		freq, err := getFrequencyMap(buf)
		if err != nil {
			t.Fatalf("unexpected error while reading from empty buffer: %v", err)
		}
		if len(freq) != 0 {
			t.Fatalf("non empty frequency map from empty readed, got: %v", freq)
		}
	})

	t.Run("broken reader", func(t *testing.T) {
		source := "aaccabbacbabac"
		reader := errorReader{limit: 5, reader: bytes.NewBufferString(source)}
		freq, err := getFrequencyMap(&reader)
		if err == nil {
			t.Fatal("expected error bot got <nil>")
		}
		if freq != nil {
			t.Fatalf("expected <nil> map but got: %v", freq)
		}
	})

	t.Run("one byte", func(t *testing.T) {
		freq, err := getFrequencyMap(bytes.NewBufferString("a"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(freq) != 1 {
			t.Fatalf("invalid frequency map size, expected: %d, got: %d", 1, len(freq))
		}
		if n, ok := freq['a']; !ok || n != 1 {
			t.Fatalf("invalid frequncy value, expected: 'a': %d, got: 'a': %d", 1, n)
		}
	})

	t.Run("several bytes", func(t *testing.T) {
		input := "aabbbccccddddd"
		freq, err := getFrequencyMap(bytes.NewBufferString(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := map[byte]uint{
			'a': 2,
			'b': 3,
			'c': 4,
			'd': 5,
		}

		if len(freq) != len(expected) {
			t.Fatalf("invalid frequency map size, expected: %d, got: %d", len(expected), len(freq))
		}

		for k, v := range expected {
			if n, ok := freq[k]; !ok || n != v {
				t.Fatalf("invalid frequency for '%c', expected: %d, got: %d", k, v, n)
			}
		}
	})
}

func FuzzGetFrequencyMap(f *testing.F) {
	testcases := []string{
		"Hello, world!",
		"",
		" ",
		"12345",
		"aaaaacccccbbbbbbbb",
		"aaacaaacccccbccbbaabbvccbbaab",
		"abcdefghijklmnopqrstuvwxyz",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
	}
	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, a string) {
		freq, err := getFrequencyMap(bytes.NewBufferString(a))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var total uint = 0
		for _, n := range freq {
			total += n
		}
		if total != uint(len(a)) {
			t.Fatalf("invalid size, expected: %d, got: %d", len(a), total)
		}
	})
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

func FuzzBuldCodesPrefix(f *testing.F) {
	testcases := []string{
		"",
		"q",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. In ac nunc accumsan",
		"aaaaaaaaaaaaaaaaaaaaaaaa",
		"aaaaaaaaaaaaaabbbbbbbbbbbbbbb",
		string([]byte{1, 234, 12, 34, 42, 142, 123}),
	}
	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, a string) {
		reader := bytes.NewBufferString(a)
		freq, err := getFrequencyMap(reader)
		if err != nil {
			t.Fatalf("unexpected error while building frequency map: %v", err)
		}
		codes := buildCodes(buildTree(freq))
		for char1, code1 := range codes {
			for char2, code2 := range codes {
				if strings.HasPrefix(code1, code2) && char1 != char2 {
					t.Fatalf(
						"prefix code invariant violated: char %q has code %q, char %q has code %q",
						char1, code1,
						char2, code2,
					)
				}
			}
		}
	})
}

func BenchmarkBuildTree(b *testing.B) {
	cases := []struct {
		name string
		data string
	}{
		{"range_1", benchkit.Range(0, 17)},
		{"range_2", benchkit.Range(0, 34)},
		{"range_3", benchkit.Range(0, 51)},
		{"range_4", benchkit.Range(0, 68)},
		{"range_5", benchkit.Range(0, 85)},
		{"range_6", benchkit.Range(0, 102)},
		{"range_7", benchkit.Range(0, 119)},
		{"range_8", benchkit.Range(0, 136)},
		{"range_9", benchkit.Range(0, 153)},
		{"range_10", benchkit.Range(0, 170)},
		{"range_11", benchkit.Range(0, 187)},
		{"range_12", benchkit.Range(0, 204)},
		{"range_13", benchkit.Range(0, 221)},
		{"range_14", benchkit.Range(0, 238)},
		{"range_15", benchkit.Range(0, 256)},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			reader := bytes.NewBufferString(tc.data)
			freq, err := getFrequencyMap(reader)
			if err != nil {
				b.Fatalf("unexpected error while building frequency map: %v", err)
			}
			for b.Loop() {
				buildTree(freq)
			}
		})
	}
}

func TestCalculateTreeSize(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if size := calculateTreeSize(nil); size != 0 {
			t.Fatalf("expected 0 but got %d", size)
		}
	})

	t.Run("one leaf", func(t *testing.T) {
		const expected = 1 + 1 + 8
		root := &node{
			left: &node{char: 'a', count: 1},
		}
		if size := calculateTreeSize(root); size != expected {
			t.Fatalf("invalud tree size, expected: %d, got: %d", expected, size)
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
		const expected = 1 + 2 + 2 + 2 + 4*8
		if size := calculateTreeSize(root); size != expected {
			t.Fatalf("invalud tree size, expected: %d, got: %d", expected, size)
		}
	})
}

func TestCalculateContentSize(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		size, err := calculateContentSize(map[byte]string{}, map[byte]uint{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if size != 0 {
			t.Fatalf("invalid size, 0 expected, got: %d", size)
		}
	})

	t.Run("default", func(t *testing.T) {
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
	})

	t.Run("error handling", func(t *testing.T) {
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
	})
}

func TestBuildReverseCodes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if codes := buildReverseCodes(nil); len(codes) != 0 {
			t.Fatalf("wrong reverse codes size, expected: %d, got: %d", 0, len(codes))
		}
	})

	// this case is considered logically unreachable during normal execution
	t.Run("single leaf, no internal", func(t *testing.T) {
		root := &node{char: 'A'}
		codes := buildReverseCodes(root)

		if len(codes) != 1 {
			t.Fatalf("expected 1 code, got %d", len(codes))
		}
		if codes[""] != 'A' {
			t.Fatalf("expected code '' -> 'A', got %v", codes)
		}
	})

	t.Run("single leaf", func(t *testing.T) {
		root := &node{
			left: &node{char: 'a'},
		}
		codes := buildReverseCodes(root)

		if len(codes) != 1 {
			t.Fatalf("expected 1 code, got %d", len(codes))
		}
		if codes["1"] != 'a' {
			t.Fatalf("incorrect codes: %v", codes)
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

		expected := map[string]byte{
			"1":   'd',
			"01":  'c',
			"001": 'a',
			"000": 'b',
		}
		codes := buildReverseCodes(root)
		if len(codes) != 4 {
			t.Fatalf("expected 4 codes, got %d", len(codes))
		}
		for k, v := range expected {
			if codes[k] != v {
				t.Fatalf("expected code %s -> %c, got %v", k, v, codes)
			}
		}
	})
}

func TestWriteCodes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if err := writeCodes(nil, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("single leaf", func(t *testing.T) {
		root := &node{
			left: &node{char: 'a', count: 10},
		}
		buf := &bytes.Buffer{}
		writer := bitio.NewWriter(buf)
		if err := writeCodes(root, writer); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		writer.Flush()
		result := buf.Bytes()
		if len(result) != 2 {
			t.Fatalf("invalid content size, expected: %d, got: %d", 2, len(result))
		}

		if result[0]>>7 != 0 {
			t.Fatal("invalid content structure, expected first bit to be 0, got 1")
		}
		if (result[0]&0b01000000)>>6 != 1 {
			t.Fatal("invalid content structure, expected second bit to be 1, got 0")
		}
		char := byte((result[0]&0b00111111)<<2 | ((result[1] & 0b11000000) >> 6))
		if char != 'a' {
			t.Fatalf("invalid content structure, next 8 bits should represent character 'a' but got: %#08b", char)
		}
		if (result[1] & 0b00111111) != 0 {
			t.Fatalf("the rest of byte should be filled with 0 bits, got: %#08b", result[1]&0b00111111)
		}
	})

	t.Run("full tree", func(t *testing.T) {
		root := &node{
			left:  &node{char: 'a', count: 10},
			right: &node{char: 'b', count: 15},
		}
		buf := &bytes.Buffer{}
		writer := bitio.NewWriter(buf)
		if err := writeCodes(root, writer); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		writer.Flush()
		result := buf.Bytes()
		if len(result) != 3 {
			t.Fatalf("invalid content size, expected: %d, got: %d", 3, len(result))
		}
		if (result[1]&0b00100000)>>5 != 1 {
			t.Fatalf("invalid content structure, expected bit to be 0, got 1")
		}
		char := byte(((result[1] & 0b00011111) << 3) | (result[2] >> 5))
		if char != 'b' {
			t.Fatalf("invalid content structure, next 8 bits should represent character 'b' but got: %#08b", char)
		}
	})
}
