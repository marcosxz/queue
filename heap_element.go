package queue

type Element struct {
	Value    interface{} // 值
	Priority int         // 优先级,越大越先出队
	index    int         // 位置
}

// 实现标准库heap接口
type HeapElements []*Element

func (pq HeapElements) Len() int {
	return len(pq)
}

func (pq HeapElements) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority
}

func (pq HeapElements) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *HeapElements) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Element)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *HeapElements) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
