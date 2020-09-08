# queue
队列

## 优先级队列
```text
    pq := NewPriorityQueue(context.Background(), 1000)
    // Push
    pq.Push(&Element{V: "a", P: 3})
    pq.Push(&Element{V: "b", P: 2})
    pq.Push(&Element{V: "c", P: 1})
    // Pop
    for {
        if elem, ok := pq.Pop(); !ok { // the pop is stop
            log.Printf("pop:stop\n")
            return
        } else {
            log.Printf("pop:%v--%d\n", elem.V, elem.P)
        }
    }
    // Output: `pop:a--3`
    // Output: `pop:b--2`
    // Output: `pop:c--1`
```

## 延迟队列
```text
    dq := NewDelayQueue(context.Background(), 1000, func() time.Time {
        return time.Now()
    })
    // Push
    dq.Push(&Delay{V: "a", D: time.Millisecond * 1})
    dq.Push(&Delay{V: "b", D: time.Millisecond * 2})
    dq.Push(&Delay{V: "c", D: time.Millisecond * 3})
    // Pop
    for {
        if delay, ok := dq.Pop(); !ok { // the pop is stop
            log.Printf("pop:stop\n")
            return
    	} else {
    	    log.Printf("pop:%v-%s-[%s]\n", delay.V, delay.D, delay.Deadline().Sub(time.Now()))
    	}
    }
    // Output: `pop:a-1ms-[-4.399µs]`
    // Output: `pop:b-2ms-[-163.84µs]`
    // Output: `pop:c-3ms-[-225.608µs]`
```
