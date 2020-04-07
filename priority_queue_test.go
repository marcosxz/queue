package queue

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

func TestNewPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue(1000)
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	// 停止Pop
	go func() {
		time.Sleep(time.Second * 5)
		cancel()
	}()

	// 生产Push
	for i := 0; i < 1000; i++ {
		ic := *&i
		go func() {
			elem := &Element{Value: ic, Priority: ic}
			pq.Push(elem)
			log.Println("push1:", elem)
		}()
	}

	// 生产Push
	for i := 0; i < 1000; i++ {
		ic := *&i
		go func() {
			elem := &Element{Value: ic, Priority: ic}
			pq.Push(elem)
			log.Println("push2:", elem)
		}()
	}

	// 消费Pop
	wg.Add(1)
	go func() {
		for {
			elem := pq.Pop(ctx)
			log.Println("pop1:", elem)
			if elem == nil { // 得到的元素为空时说明队列已停止
				log.Println("pop1:stop")
				wg.Done()
				return
			}
		}
	}()

	// 消费Pop
	wg.Add(1)
	go func() {
		for {
			elem := pq.Pop(ctx)
			log.Println("pop2:", elem)
			if elem == nil {
				log.Println("pop2:stop")
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
}
