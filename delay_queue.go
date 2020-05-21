package queue

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// 优先级队列
type DelayQueue struct {
	sync.RWMutex
	h        heap.Interface
	state    int32
	wakeC    chan struct{}
	ctx      context.Context
	cancel   context.CancelFunc
	minDelay time.Duration
}

// Len队列长度
func (pq *DelayQueue) Len() int {
	pq.RLock()
	n := pq.h.Len()
	pq.RUnlock()
	return n
}

// Push入队
func (pq *DelayQueue) Push(elem *Delay) {
	if elem != nil {
		elem.deadline = time.Now().Add(elem.Delay)
		pq.Lock()
		heap.Push(pq.h, elem)
		pq.Unlock()
		if atomic.CompareAndSwapInt32(&pq.state, blocking, normal) {
			pq.wakeC <- struct{}{}
		}
	}
}

// Pop出队
func (pq *DelayQueue) Pop() *Delay {
	for {
		var elem *Delay
		pq.Lock()
		if n := pq.h.Len(); n > 0 {
			hi := *pq.h.(*HeapDelays)
			elem = hi[0]
			if d := elem.deadline.Sub(time.Now()); d > 0 {
				after := pq.minDelay
				if after > d {
					after = d
				}
				<-time.After(after)
				pq.Unlock()
				continue
			}
			heap.Remove(pq.h, elem.index)
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
func (pq *DelayQueue) Fix(elem *Delay) {
	if elem != nil && elem.index > -1 {
		pq.Lock()
		heap.Fix(pq.h, elem.index)
		pq.Unlock()
	}
}

// Remove删除队列中的元素
func (pq *DelayQueue) Remove(elem *Delay) {
	if elem != nil && elem.index > -1 {
		pq.Lock()
		heap.Remove(pq.h, elem.index)
		pq.Unlock()
	}
}

// Peek查看但不出队
func (pq *DelayQueue) Peek(idx int) *Delay {
	if idx < 0 {
		return nil
	}
	var elem *Delay
	pq.RLock()
	if n := pq.h.Len(); n > idx {
		hi := *pq.h.(*HeapDelays)
		elem = hi[idx]
	}
	pq.RUnlock()
	return elem
}

func (pq *DelayQueue) Stop() {
	pq.cancel()
}

func NewDelayQueue(minDelay time.Duration, cap int) *DelayQueue {
	h := make(HeapDelays, 0, cap)
	heap.Init(&h)
	ctx, cancel := context.WithCancel(context.Background())
	return &DelayQueue{h: &h, ctx: ctx, cancel: cancel, minDelay: minDelay, wakeC: make(chan struct{})}
}
