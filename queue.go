package queue2

// 队列接口,该包的所有队列都应实现该接口
type Queue interface {
	// Len队列长度
	Len() int
	// Pop出队一个元素
	Pop() *Element
	// Push入队一个元素
	Push(*Element)
	// Peek查看指定位置的一个元素,但不出队
	Peek(int) *Element
	// Fix更新该元素并根据优先级重新固定位置
	Fix(*Element)
	// Remove删除队列中的该元素
	Remove(*Element)
	// Off关闭队列
	Off()
}

// 队列元素
type Element struct {
	// 元素值
	Value interface{}
	// 优先级,越大越先出队
	Priority int
	// 元素位置
	index int
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
