package queue

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNewPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue(1000)
	wg := &sync.WaitGroup{}

	// 停止
	go func() {
		time.Sleep(time.Second * 5)
		pq.Off()
	}()

	// 生产1
	for i := 0; i < 100; i++ {
		wg.Add(1)
		ic := *&i
		go func() {
			elem := &Element{Value: strconv.Itoa(ic), Priority: ic}
			pq.Push(elem)
			t.Log("push1:", elem)
			wg.Done()
		}()
	}

	// 消费1
	wg.Add(1)
	go func() {
		for {
			elem := pq.Pop()
			t.Log("pop1:", elem)
			if elem == nil { // 得到的元素为空时说明队列已停止
				t.Log("pop1:stop")
				wg.Done()
				return
			}
		}
	}()

	// 消费2
	wg.Add(1)
	go func() {
		for {
			elem := pq.Pop()
			t.Log("pop2:", elem)
			if elem == nil {
				t.Log("pop2:stop")
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
}
