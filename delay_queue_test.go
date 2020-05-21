package queue

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestNewDelayQueue(t *testing.T) {
	pq := NewDelayQueue(time.Second, 1000)
	wg := &sync.WaitGroup{}

	// 停止Pop
	go func() {
		time.Sleep(time.Second * 60)
		pq.Stop()
	}()

	// 生产Push
	for i := 0; i < 1000; i++ {
		ic := *&i
		go func() {
			elem := &Delay{Value: ic, Delay: time.Millisecond * time.Duration(ic)}
			pq.Push(elem)
			//log.Println("push1:", elem)
		}()
	}

	// 生产Push
	go func() {
		for i := 0; i < 1000; i++ {
			ic := *&i
			go func() {
				elem := &Delay{Value: ic, Delay: time.Millisecond * time.Duration(ic)}
				pq.Push(elem)
				//log.Println("push2:", elem)
			}()
		}
	}()

	// 消费Pop
	wg.Add(1)
	go func() {
		for {
			elem := pq.Pop()
			log.Println("pop1:", elem, elem.deadline.Sub(time.Now()))
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
			log.Println("pop2:", elem, elem.deadline.Sub(time.Now()))
			if elem == nil {
				log.Println("pop2:stop", pq.Len())
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
}
