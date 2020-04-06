package queue

import (
	"sync"
	"testing"
	"time"
)

func TestNewDelayQueue(t *testing.T) {
	dq := NewDelayQueue(10)
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		ic := *&i
		go func() {
			dq.Push(ic, time.Second*time.Duration(ic))
			wg.Done()
		}()
	}

	go func() {
		time.Sleep(time.Second * 10)
		dq.Off()
	}()

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
