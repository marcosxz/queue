# queue
队列

## 优先级队列
```text
	pq := NewPriorityQueue(1000)
	
	// 生产
	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(time.Second)
			elem := &Element{Value: strconv.Itoa(i), Priority: i}
			pq.Push(elem)
			t.Log("push:", elem)
		}
	}()
	
	// 停止
	go func() {
		time.Sleep(time.Second * 10)
		pq.Off()
	}()
	
	// 消费
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			elem := pq.Pop()
			t.Log("pop:", elem)
			if elem == nil { // 得到的元素为空时说明队列已停止
				t.Log("pop:stop")
				wg.Done()
				return
			}
		}
	}()
	
	wg.Wait()
```
