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
	//ctx, cancel := context.WithCancel(context.Background())

	// 停止Pop
	//go func() {
	//	time.Sleep(time.Second * 60)
	//	pq.Off()
	//}()

	// 生产Push
	for i := 0; i < 100; i++ {
		wg.Add(1)
		ic := *&i
		go func() {
			time.Sleep(time.Second * time.Duration(ic))
			elem := &Element{Value: strconv.Itoa(ic), Priority: ic}
			pq.Push(elem)
			t.Log("push:", elem)
			wg.Done()
		}()
	}

	// 消费Pop
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

	// 消费Pop
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
