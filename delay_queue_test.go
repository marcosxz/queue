package queue

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

func TestNewDelayQueue(t *testing.T) {
	dq := NewDelayQueue(10)
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < 100; i++ {
		wg.Add(1)
		ic := *&i
		go func() {
			dq.Push(&Delay{Value: ic, Delay: time.Second * time.Duration(ic)})
			wg.Done()
		}()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		ic := *&i
		go func() {
			dq.Push(&Delay{Value: ic, Delay: time.Second * time.Duration(ic)})
			wg.Done()
		}()
	}

	go func() {
		time.Sleep(time.Second * 10)
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			val := dq.Pop(ctx)
			log.Println("pop:", val)
			if val == nil {
				log.Println("pop:stop")
				return
			}
		}
	}()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	for {
	// 		val := dq.Pop(ctx)
	// 		log.Println("pop2:", val)
	// 		if val == nil {
	// 			log.Println("pop2:stop")
	// 			return
	// 		}
	// 	}
	// }()

	wg.Wait()
}
