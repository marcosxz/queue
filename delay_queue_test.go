package queue

import (
	"sync"
	"testing"
	"time"
)

func TestNewDelayQueue(t *testing.T) {
	dq := NewDelayQueue(10)
	dq.Push(1, time.Second*1)
	dq.Push(2, time.Second*2)
	dq.Push(3, time.Second*3)
	dq.Push(4, time.Second*4)
	//dq.Push(8, time.Second*8)

	go func() {
		time.Sleep(time.Second * 10)
		dq.Off()
	}()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			val := dq.Pop()
			t.Log("pop:", val)
			if val == nil {
				t.Log("pop:stop")
				return
			}
		}
	}()

	wg.Wait()
}
