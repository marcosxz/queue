package queue

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

var ec, sc, ecMux, scMux = 0, 0, &sync.Mutex{}, &sync.Mutex{}

func incrEc() {
	ecMux.Lock()
	defer ecMux.Unlock()
	ec++
}

func incrSc() {
	scMux.Lock()
	defer scMux.Unlock()
	sc++
}

func TestPriorityQueuePushPop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pq := NewPriorityQueue(ctx, 1000)
	wg := &sync.WaitGroup{}

	go func() {
		time.Sleep(time.Second * 100)
		cancel()
	}()

	// Push
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			pq.Push(&Element{V: i, P: i})
			wg.Done()
		}(i)
	}

	// Pop
	wg.Add(1)
	go func() {
		for {
			if elem, ok := pq.Pop(); !ok { // the pop is stop
				log.Printf("pop:stop\n")
				incrSc()
				wg.Done()
				return
			} else {
				incrEc()
				log.Printf("pop:%d--%d\n", elem.V, elem.P)
			}
		}
	}()

	// wait
	wg.Wait()

	// check pass
	if ec == 1000 && sc == 1 {
		t.Log("OK")
	} else {
		t.Errorf("ec:%d, sc:%d", ec, sc)
	}
}

func TestPriorityQueuePeekFixRemove(t *testing.T) {
	pq := NewPriorityQueue(context.Background(), 1000)

	// Push
	for i := 0; i < 500; i++ {
		pq.Push(&Element{V: i, P: i})
	}

	// Peek Fix
	isFx := true
	for i := 0; i < 500; i++ {
		elem := pq.Peek(0)
		if elem != nil {
			pre := elem.index
			elem.P = -1
			pq.Fix(elem)
			if elem.index != 0 && pre == elem.index {
				isFx = false
			}
		}
	}
	if isFx {
		t.Log("peek fix ok")
	} else {
		t.Error("peek fix fail")
		return
	}

	// Peek Remove
	isRm := true
	for i := 0; i < 500; i++ {
		elem := pq.Peek(0)
		if elem != nil {
			rmd := pq.Remove(elem)
			if rmd.V != elem.V {
				isRm = false
				break
			}
		}
	}
	if isRm && pq.Len() == 0 {
		t.Log("peek remove ok")
	} else {
		t.Error("peek remove fail")
		return
	}
}
