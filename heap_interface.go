package queue

import (
	"container/heap"
	"time"
)

type Element struct {
	V     interface{} // value
	P     int         // priority: max is first pop
	index int         // heap index
}

func (e *Element) Index() int {
	return e.index
}

// impl the std lib heap interface
type HeapElements struct {
	cap int
	e   []*Element
}

func NewHeapElements(cap int) heap.Interface {
	return &HeapElements{cap: cap, e: make([]*Element, 0, cap)}
}

func (he *HeapElements) Len() int {
	return len(he.e)
}

func (he *HeapElements) Less(i, j int) bool {
	return he.e[i].P > he.e[j].P
}

func (he *HeapElements) Swap(i, j int) {
	he.e[i], he.e[j] = he.e[j], he.e[i]
	he.e[i].index, he.e[j].index = i, j
}

func (he *HeapElements) Push(x interface{}) {
	n := len(he.e)
	c := cap(he.e)
	if n >= c {
		ne := make([]*Element, n, c*2)
		copy(ne, he.e)
		he.e = ne
	}
	item := x.(*Element)
	item.index = n
	he.e = append(he.e, item)
}

func (he *HeapElements) Pop() interface{} {
	n := len(he.e)
	c := cap(he.e)
	half := c / 2
	if n < half && half >= he.cap {
		ne := make([]*Element, n, half)
		copy(ne, he.e)
		he.e = ne
	}
	item := he.e[n-1]
	item.index = -1
	he.e[n-1] = nil
	he.e = he.e[0 : n-1]
	return item
}

// impl the std lib heap interface
type Delay struct {
	V        interface{}   // value
	D        time.Duration // delay: min time.Millisecond is better
	deadline time.Time
	index    int
}

func (d *Delay) Deadline() time.Time {
	return d.deadline
}

func (d *Delay) Index() int {
	return d.index
}

type HeapDelays struct {
	cap int
	e   []*Delay
}

func NewHeapDelays(cap int) heap.Interface {
	return &HeapDelays{cap: cap, e: make([]*Delay, 0, cap)}
}

func (hd *HeapDelays) Len() int {
	return len(hd.e)
}

func (hd *HeapDelays) Less(i, j int) bool {
	return hd.e[i].D < hd.e[j].D
}

func (hd *HeapDelays) Swap(i, j int) {
	hd.e[i], hd.e[j] = hd.e[j], hd.e[i]
	hd.e[i].index, hd.e[j].index = i, j
}

func (hd *HeapDelays) Push(x interface{}) {
	n := len(hd.e)
	c := cap(hd.e)
	if n >= c {
		ne := make([]*Delay, n, c*2)
		copy(ne, hd.e)
		hd.e = ne
	}
	item := x.(*Delay)
	item.index = n
	hd.e = append(hd.e, item)
}

func (hd *HeapDelays) Pop() interface{} {
	n := len(hd.e)
	c := cap(hd.e)
	half := c / 2
	if n < half && half >= hd.cap {
		ne := make([]*Delay, n, half)
		copy(ne, hd.e)
		hd.e = ne
	}
	item := hd.e[n-1]
	item.index = -1
	hd.e[n-1] = nil
	hd.e = hd.e[0 : n-1]
	return item
}
