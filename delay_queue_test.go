package queue

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"
)

func init() {
	go func() {
		http.ListenAndServe(":8080", nil)
	}()
}

func TestDelayQueuePushPop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	dq := NewDelayQueue(ctx, 1000, func() time.Time {
		return time.Now()
	})
	wg := &sync.WaitGroup{}

	// Stop
	go func() {
		time.Sleep(time.Second * 1)
		cancel()
	}()

	// Pop
	wg.Add(1)
	go func() {
		for {
			if delay, ok := dq.Pop(); !ok { // the pop is stop
				log.Printf("pop:stop\n")
				wg.Done()
				return
			} else {
				log.Printf("pop:%d-%s-%s\n", delay.V, delay.D, delay.deadline.Sub(time.Now()))
			}
		}
	}()

	// Push
	go func() {
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func(i int) {
				dq.Push(&Delay{V: i, D: time.Millisecond * time.Duration(i)})
				wg.Done()
			}(i)
		}
	}()

	wg.Wait()
}
