package queue

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const (
	expired  = 0
	expiring = 1
	sleeping = 2
)

// 延迟队列
type DelayQueue struct {
	sync.RWMutex
	h      heap.Interface
	delay  time.Duration
	sleep  int32
	delayC chan struct{}
	state  int32
	wakeC  chan struct{}
}

// Push入队一个
func (dq *DelayQueue) Push(delay *Delay) {
	delay.deadline = time.Now().Add(delay.Delay)
	change := false
	dq.Lock()
	heap.Push(dq.h, delay)
	if dq.delay == 0 || dq.delay > delay.Delay {
		dq.delay = delay.Delay
		change = true
	}
	dq.Unlock()
	if change && atomic.CompareAndSwapInt32(&dq.sleep, sleeping, expired) {
		dq.delayC <- struct{}{}
	}
	if atomic.CompareAndSwapInt32(&dq.state, blocking, normal) {
		dq.wakeC <- struct{}{}
	}
}

// Pop满足指定条件的元素才出队
func (dq *DelayQueue) Pop(ctx context.Context) *Delay {
	for {
		var delay *Delay
		var delta time.Duration
		dq.Lock()
		if n := dq.h.Len(); n > 0 {
			hi := *dq.h.(*HeapDelays)
			delay = hi[0]
			delta = time.Now().Sub(delay.deadline)
			if delta <= 0 {
				atomic.StoreInt32(&dq.sleep, expiring)
			} else {
				heap.Remove(dq.h, delay.index)
			}
		} else {
			atomic.StoreInt32(&dq.state, waiting)
		}
		dq.Unlock()
		switch {
		case atomic.CompareAndSwapInt32(&dq.state, waiting, blocking):
			select {
			case <-dq.wakeC:
				continue
			case <-ctx.Done():
				atomic.StoreInt32(&dq.state, normal)
				return nil
			}
		default:
			switch {
			case atomic.CompareAndSwapInt32(&dq.sleep, expiring, sleeping):
			reset:
				select {
				case <-time.After(dq.delay):
				case <-ctx.Done():
					atomic.StoreInt32(&dq.sleep, expired)
					return nil
				case <-dq.delayC:
					goto reset
				}
			default:
				select {
				case <-ctx.Done():
					return nil
				default:
					return delay
				}
			}
		}
	}
}

func NewDelayQueue(cap int) *DelayQueue {
	h := make(HeapDelays, 0, cap)
	heap.Init(&h)
	return &DelayQueue{h: &h, wakeC: make(chan struct{}), delayC: make(chan struct{})}
}
