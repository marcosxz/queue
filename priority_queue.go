package queue

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
)

const (
	normal   int32 = 0
	waiting  int32 = 1
	blocking int32 = 2
)

// 优先级队列
type PriorityQueue struct {
	sync.RWMutex
	h      heap.Interface
	state  int32
	wakeC  chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

// Len队列长度
func (pq *PriorityQueue) Len() int {
	pq.RLock()
	n := pq.h.Len()
	pq.RUnlock()
	return n
}

// Push入队
func (pq *PriorityQueue) Push(elem *Element) {
	if elem != nil {
		pq.Lock()
		heap.Push(pq.h, elem)
		pq.Unlock()
		if atomic.CompareAndSwapInt32(&pq.state, blocking, normal) {
			pq.wakeC <- struct{}{}
		}
	}
}

// Pop出队
func (pq *PriorityQueue) Pop() *Element {
	for {
		var elem *Element
		pq.Lock()
		if n := pq.h.Len(); n > 0 {
			elem = heap.Pop(pq.h).(*Element)
		} else {
			atomic.StoreInt32(&pq.state, waiting)
		}
		pq.Unlock()
		switch {
		case atomic.CompareAndSwapInt32(&pq.state, waiting, blocking):
			select {
			case i := <-pq.wakeC:
				select {
				case pq.wakeC <- i:
				default:
					continue
				}
			case <-pq.ctx.Done():
				return nil
			}
		default:
			select {
			case <-pq.ctx.Done():
				return nil
			default:
				return elem
			}
		}
	}
}

// Fix更新元素并根据优先级重新固定位置
func (pq *PriorityQueue) Fix(elem *Element) {
	if elem != nil && elem.index > -1 {
		pq.Lock()
		heap.Fix(pq.h, elem.index)
		pq.Unlock()
	}
}

// Remove删除队列中的元素
func (pq *PriorityQueue) Remove(elem *Element) {
	if elem != nil && elem.index > -1 {
		pq.Lock()
		heap.Remove(pq.h, elem.index)
		pq.Unlock()
	}
}

// Peek查看但不出队
func (pq *PriorityQueue) Peek(idx int) *Element {
	if idx < 0 {
		return nil
	}
	var elem *Element
	pq.RLock()
	if n := pq.h.Len(); n > idx {
		hi := *pq.h.(*HeapElements)
		elem = hi[idx]
	}
	pq.RUnlock()
	return elem
}

func (pq *PriorityQueue) Stop() {
	pq.cancel()
}

func NewPriorityQueue(cap int) *PriorityQueue {
	h := make(HeapElements, 0, cap)
	heap.Init(&h)
	ctx, cancel := context.WithCancel(context.Background())
	return &PriorityQueue{h: &h, ctx: ctx, cancel: cancel, wakeC: make(chan struct{})}
}
