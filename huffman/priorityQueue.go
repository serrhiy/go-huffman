package huffman

type priorityQueue []*node

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].count < pq[j].count
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(item any) {
	*pq = append(*pq, item.(*node))
}

func (pq *priorityQueue) Pop() any {
	item := (*pq)[len(*pq)-1]
	*pq = (*pq)[0 : len(*pq)-1]
	return item
}
