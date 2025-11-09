package delayqueue

type Item struct {
	value    any
	priority int64
	index    int
}

type priorityQueue []*Item

func (pq priorityQueue) Less(i, j int) bool { return pq[i].priority < pq[j].priority }
func (pq priorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i]; pq[i].index = i; pq[j].index = j }
func (pq priorityQueue) Len() int           { return len(pq) }
func (pq *priorityQueue) Push(x any) {
	// manully expand slice cap
	l, c := len(*pq), cap(*pq)
	if l == c {
		npq := make(priorityQueue, l, c*2)
		copy(npq, *pq)
		*pq = npq
	}
	*pq = (*pq)[:l+1]
	(*pq)[l] = x.(*Item)
	(*pq)[l].index = l
}
func (pq *priorityQueue) Pop() any {
	// manylly extract slice cap
	l, c := len(*pq), cap(*pq)
	if l < c/2 && c > 25 {
		npq := make(priorityQueue, l, c/2)
		copy(npq, (*pq)[:l])
		*pq = npq
	}
	item := (*pq)[l-1]
	item.index = -1 // for safety
	*pq = (*pq)[:l-1]
	return item
}
