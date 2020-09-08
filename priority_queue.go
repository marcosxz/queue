package queue

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
)

type PriorityQueue struct {
	sync.Mutex
	h      heap.Interface
	state  int32 // 0 none 1 sleeping
	wakeC  chan struct{}
	queueC chan *Element
}

func NewPriorityQueue(ctx context.Context, cap int) *PriorityQueue {
	pq := &PriorityQueue{
		h:      NewHeapElements(cap),
		wakeC:  make(chan struct{}),
		queueC: make(chan *Element, cap),
	}
	heap.Init(pq.h)
	go pq.pop(ctx)
	return pq
}

func (pq *PriorityQueue) Len() int {
	pq.Lock()
	n := pq.h.Len()
	pq.Unlock()
	return n
}

func (pq *PriorityQueue) Queue() <-chan *Element {
	return pq.queueC
}

func (pq *PriorityQueue) Pop() (*Element, bool) {
	elem, ok := <-pq.queueC
	return elem, ok
}

func (pq *PriorityQueue) Push(elem *Element) {
	if elem != nil {
		pq.Lock()
		heap.Push(pq.h, elem)
		idx := elem.index
		pq.Unlock()
		if idx == 0 && atomic.CompareAndSwapInt32(&pq.state, 1, 0) {
			pq.wakeC <- struct{}{}
		}
	}
}

func (pq *PriorityQueue) Fix(elem *Element) {
	pq.Lock()
	if elem != nil && elem.index > -1 {
		heap.Fix(pq.h, elem.index)
	}
	pq.Unlock()
}

func (pq *PriorityQueue) Remove(elem *Element) *Element {
	pq.Lock()
	if elem != nil && elem.index > -1 {
		if rmd := heap.Remove(pq.h, elem.index); rmd != nil {
			elem = rmd.(*Element)
		}
	}
	pq.Unlock()
	return elem
}

func (pq *PriorityQueue) Peek(idx int) *Element {
	var elem *Element
	pq.Lock()
	if n := pq.h.Len(); n > idx {
		hi := *pq.h.(*HeapElements)
		elem = hi.e[idx]
	}
	pq.Unlock()
	return elem
}

func (pq *PriorityQueue) pop(ctx context.Context) {
loop:
	var elem *Element
	pq.Lock()
	if pq.h.Len() > 0 {
		elem = heap.Pop(pq.h).(*Element)
	}
	pq.Unlock()
	if elem != nil {
		select {
		case <-ctx.Done():
			goto exit
		case pq.queueC <- elem:
			goto loop
		}
	}
	atomic.StoreInt32(&pq.state, 1)
	select {
	case s := <-pq.wakeC:
		select {
		case pq.wakeC <- s: // notify other sleep wakeC
			goto loop
		default:
			goto loop
		}
	case <-ctx.Done():
		goto exit
	}
exit:
	close(pq.queueC)
	atomic.StoreInt32(&pq.state, 0)
}
