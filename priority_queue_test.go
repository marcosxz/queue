package queue

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestNewPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue(1000)
	wg := &sync.WaitGroup{}

	// 停止Pop
	go func() {
		time.Sleep(time.Second * 30)
		pq.Stop()
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
	go func() {
		time.Sleep(time.Second * 2)
		for i := 0; i < 1000; i++ {
			ic := *&i
			go func() {
				elem := &Element{Value: ic, Priority: ic}
				pq.Push(elem)
				log.Println("push2:", elem)
			}()
		}
	}()

	// 消费Pop
	wg.Add(1)
	go func() {
		for {
			elem := pq.Pop()
			log.Println("pop1:", elem)
			if elem == nil { // 得到的元素为空时说明队列已停止
				log.Println("pop1:stop", pq.Len())
				wg.Done()
				return
			}
		}
	}()

	// 消费Pop
	wg.Add(1)
	go func() {
		for {
			elem := pq.Pop()
			log.Println("pop2:", elem)
			if elem == nil {
				log.Println("pop2:stop", pq.Len())
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
}
