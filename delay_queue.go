package queue

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type DelayQueue struct {
	sync.Mutex
	h      heap.Interface
	state  int32 // 0 none 1 sleeping
	wakeC  chan struct{}
	queueC chan *Delay
	nowFn  func() time.Time
}

func NewDelayQueue(ctx context.Context, cap int, now func() time.Time) *DelayQueue {
	if now == nil {
		now = func() time.Time { return time.Now() }
	}
	dq := &DelayQueue{
		nowFn:  now,
		h:      NewHeapDelays(cap),
		wakeC:  make(chan struct{}),
		queueC: make(chan *Delay, cap),
	}
	heap.Init(dq.h)
	go dq.pop(ctx)
	return dq
}

func (dq *DelayQueue) Len() int {
	dq.Lock()
	n := dq.h.Len()
	dq.Unlock()
	return n
}

func (dq *DelayQueue) Queue() <-chan *Delay {
	return dq.queueC
}

func (dq *DelayQueue) Pop() (*Delay, bool) {
	delay, ok := <-dq.queueC
	return delay, ok
}

func (dq *DelayQueue) Push(delay *Delay) {
	if delay != nil {
		delay.deadline = dq.nowFn().Add(delay.D)
		dq.Lock()
		heap.Push(dq.h, delay)
		idx := delay.index
		dq.Unlock()
		if idx == 0 && atomic.CompareAndSwapInt32(&dq.state, 1, 0) {
			dq.wakeC <- struct{}{}
		}
	}
}

func (dq *DelayQueue) Fix(delay *Delay) {
	dq.Lock()
	if delay != nil && delay.index > -1 {
		heap.Fix(dq.h, delay.index)
	}
	dq.Unlock()
}

func (dq *DelayQueue) Remove(delay *Delay) *Delay {
	dq.Lock()
	if delay != nil && delay.index > -1 {
		if rmd := heap.Remove(dq.h, delay.index); rmd != nil {
			delay = rmd.(*Delay)
		}
	}
	dq.Unlock()
	return delay
}

func (dq *DelayQueue) Peek(idx int) *Delay {
	var delay *Delay
	dq.Lock()
	if n := dq.h.Len(); n > idx {
		hi := *dq.h.(*HeapDelays)
		delay = hi.e[idx]
	}
	dq.Unlock()
	return delay
}

func (dq *DelayQueue) pop(ctx context.Context) {
	timer := time.NewTimer(0)
loop:
	dq.Lock()
	removed, remaining, ok := dq.expiredRemove()
	dq.Unlock()
	if remaining > 0 {
		timer.Reset(remaining)
	}
	switch {
	case !ok: // queue is empty
		atomic.StoreInt32(&dq.state, 1)
		select {
		case <-dq.wakeC:
			goto loop
		case <-ctx.Done():
			goto exit
		}
	case removed == nil: // no expired item
		atomic.StoreInt32(&dq.state, 1)
		select {
		case <-dq.wakeC:
			goto loop
		case <-ctx.Done():
			goto exit
		case <-timer.C:
			select { // the dq.wakeC is continue, unblock the Push
			case <-dq.wakeC:
				goto loop
			default:
				goto loop
			}
		}
	default: // item is expired and is removed
		select {
		case <-ctx.Done():
			goto exit
		case dq.queueC <- removed:
			goto loop
		}
	}
exit:
	close(dq.queueC)
	atomic.StoreInt32(&dq.state, 0)
}

func (dq *DelayQueue) expiredRemove() (*Delay, time.Duration, bool) {
	if dq.h.Len() == 0 {
		return nil, 0, false
	}
	hi := *dq.h.(*HeapDelays)
	delay := hi.e[0]
	if diff := delay.deadline.Sub(dq.nowFn()); diff > 0 {
		return nil, diff, true
	}
	heap.Remove(dq.h, delay.index)
	return delay, 0, true
}
