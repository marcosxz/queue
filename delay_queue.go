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
	h          heap.Interface
	ctx        context.Context
	cancel     context.CancelFunc
	delay      time.Duration
	delayState int32
	delayWake  chan struct{}
	pushState  int32
	pushWake   chan struct{}
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
	change := false
	dq.Lock()
	heap.Push(dq.h, elem)
	if dq.delay == 0 || delay < dq.delay {
		dq.delay = delay
		change = true
	}
	dq.Unlock()
	if change && atomic.CompareAndSwapInt32(&dq.delayState, blocking, normal) {
		dq.delayWake <- struct{}{}
	}
	if atomic.CompareAndSwapInt32(&dq.pushState, blocking, normal) {
		dq.pushWake <- struct{}{}
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
			elem = hi[0]
			if delta = time.Now().UnixNano() + int64(elem.Priority); delta >= 0 {
				heap.Pop(dq.h)
			} else {
				atomic.CompareAndSwapInt32(&dq.pushState, normal, waiting)
			}
		} else {
			atomic.CompareAndSwapInt32(&dq.pushState, normal, waiting)
		}
		dq.RUnlock()
		switch {
		case atomic.CompareAndSwapInt32(&dq.pushState, waiting, blocking):
			select {
			case <-dq.pushWake:
				continue
			case <-dq.ctx.Done():
				return nil
			}
		default:
			switch {
			case atomic.CompareAndSwapInt32(&dq.pushState, waiting, blocking):
			rest:
				select {
				case <-time.After(dq.delay):
				case <-dq.delayWake:
					goto rest
				}
			default:
				return elem.Value
			}
		}
	}
}

func (dq *DelayQueue) Off() { dq.cancel() }

func NewDelayQueue(cap int) *DelayQueue {
	h := make(HeapElements, 0, cap)
	heap.Init(&h)
	ctx, cancel := context.WithCancel(context.Background())
	return &DelayQueue{h: &h, ctx: ctx, cancel: cancel, delayWake: make(chan struct{}, 1), pushWake: make(chan struct{}, 1)}
}
