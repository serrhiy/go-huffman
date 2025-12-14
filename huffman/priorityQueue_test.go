package huffman

import (
	"container/heap"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var queue priorityQueue
		heap.Init(&queue)

		if queue.Len() != 0 {
			t.Fatalf("empty que has non zero length: %d", queue.Len())
		}
	})

	t.Run("len", func(t *testing.T) {
		var queue priorityQueue
		heap.Init(&queue)

		if queue.Len() != 0 {
			t.Fatalf("expected length 0, got %d", queue.Len())
		}

		heap.Push(&queue, &node{char: 'a', count: 1})
		if queue.Len() != 1 {
			t.Fatalf("expected length 1, got %d", queue.Len())
		}
	})

	t.Run("default arguments", func(t *testing.T) {
		queue := priorityQueue{
			&node{char: 'a', count: 3},
			&node{char: 'b', count: 1},
			&node{char: 'c', count: 2},
		}

		heap.Init(&queue)

		if n := heap.Pop(&queue).(*node); n.count != 1 {
			t.Fatalf("expected min count 1, got %d", n.count)
		}
		if n := heap.Pop(&queue).(*node); n.count != 2 {
			t.Fatalf("expected min count 2, got %d", n.count)
		}
		if n := heap.Pop(&queue).(*node); n.count != 3 {
			t.Fatalf("expected min count 3, got %d", n.count)
		}
	})

	t.Run("push", func(t *testing.T) {
		var queue priorityQueue
		heap.Init(&queue)

		heap.Push(&queue, &node{char: 'a', count: 5})
		heap.Push(&queue, &node{char: 'b', count: 2})
		heap.Push(&queue, &node{char: 'c', count: 8})

		if queue.Len() != 3 {
			t.Fatalf("expected length 3, got %d", queue.Len())
		}

		if n := heap.Pop(&queue).(*node); n.char != 'b' || n.count != 2 {
			t.Fatalf("expected {b,2}, got {%c,%d}", n.char, n.count)
		}
		if n := heap.Pop(&queue).(*node); n.char != 'a' || n.count != 5 {
			t.Fatalf("expected {b,2}, got {%c,%d}", n.char, n.count)
		}
		if n := heap.Pop(&queue).(*node); n.char != 'c' || n.count != 8 {
			t.Fatalf("expected {b,2}, got {%c,%d}", n.char, n.count)
		}
	})

	t.Run("pop", func(t *testing.T) {
		queue := priorityQueue{
			&node{char: 'a', count: 4},
			&node{char: 'b', count: 1},
			&node{char: 'c', count: 3},
		}
		heap.Init(&queue)

		expected := []uint{1, 3, 4}

		for i, exp := range expected {
			n := heap.Pop(&queue).(*node)
			if n.count != exp {
				t.Fatalf("pop %d: expected %d, got %d", i, exp, n.count)
			}
		}

		if queue.Len() != 0 {
			t.Fatalf("queue should be empty")
		}
	})

	t.Run("swap", func(t *testing.T) {
		queue := priorityQueue{
			&node{char: 'a', count: 1},
			&node{char: 'b', count: 2},
		}

		queue.Swap(0, 1)

		if queue[0].char != 'b' || queue[1].char != 'a' {
			t.Fatalf("swap failed")
		}
	})

	t.Run("push pop", func(t *testing.T) {
		var pq priorityQueue
		heap.Init(&pq)

		heap.Push(&pq, &node{char: 'a', count: 3})
		heap.Push(&pq, &node{char: 'b', count: 1})
		heap.Push(&pq, &node{char: 'c', count: 2})

		if n := heap.Pop(&pq).(*node); n.char != 'b' || n.count != 1 {
			t.Fatalf("expected first pop {b,1}, got {%c,%d}", n.char, n.count)
		}

		if n := heap.Pop(&pq).(*node); n.char != 'c' || n.count != 2 {
			t.Fatalf("expected second pop {c,2}, got {%c,%d}", n.char, n.count)
		}

		if n := heap.Pop(&pq).(*node); n.char != 'a' || n.count != 3 {
			t.Fatalf("expected third pop {a,3}, got {%c,%d}", n.char, n.count)
		}

		if pq.Len() != 0 {
			t.Fatalf("expected empty queue, got %d", pq.Len())
		}
	})

	t.Run("pop from empty queue panics", func(t *testing.T) {
		var pq priorityQueue
		heap.Init(&pq)

		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic when popping from empty queue")
			}
		}()

		heap.Pop(&pq)
	})
}
