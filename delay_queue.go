package queue

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// 延迟队列
type DelayQueue struct {
	sync.RWMutex
	h      heap.Interface
	ctx    context.Context
	cancel context.CancelFunc
	wait   int32
	wake   chan struct{}
}

// Len队列长度
func (dq *DelayQueue) Len() int {
	dq.RLock()
	n := dq.h.Len()
	dq.RUnlock()
	return n
}

// Push入队一个
func (dq *DelayQueue) Push(value interface{}, delay time.Duration) {
	elem := &Element{Value: value, Priority: -int(time.Now().Add(delay).UnixNano())}
	dq.Lock()
	heap.Push(dq.h, elem)
	dq.Unlock()
	if atomic.CompareAndSwapInt32(&dq.wait, 1, 0) {
		dq.wake <- struct{}{}
	}
}

// Pop满足指定条件的元素才出队
func (dq *DelayQueue) Pop() interface{} {
	for {
		var elem *Element
		var delta int64
		dq.RLock()
		if n := dq.h.Len(); n > 0 {
			hi := *dq.h.(*HeapElements)
			this := hi[0]
			delta = time.Now().UnixNano() + int64(this.Priority)
			if delta >= 0 { // 过期
				elem = heap.Pop(dq.h).(*Element)
				dq.RUnlock()
				return elem.Value
			}
		} else {
			atomic.StoreInt32(&dq.wait, 1)
		}
		dq.RUnlock()
		switch {
		case atomic.LoadInt32(&dq.wait) == 1:
			select {
			case <-dq.wake:
				continue
			case <-dq.ctx.Done():
				return nil
			}
		}
	}
}

func (dq *DelayQueue) Off() { dq.cancel() }

func NewDelayQueue(cap int) *DelayQueue {
	h := make(HeapElements, 0, cap)
	heap.Init(&h)
	ctx, cancel := context.WithCancel(context.Background())
	return &DelayQueue{h: &h, ctx: ctx, cancel: cancel, wake: make(chan struct{})}
}
