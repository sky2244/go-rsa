package main

// https://stackoverflow.com/questions/31060023/go-wait-for-next-item-in-a-priority-queue-if-empty

import (
	"sync"

	"container/list"
)

type Heap struct {
	b *list.List
	c *sync.Cond
}

func NewHeap() *Heap {
	return &Heap{
		b: list.New(),
		c: sync.NewCond(new(sync.Mutex)),
	}
}

// Pop (waits until anything available)
func (h *Heap) Pop() any {
	h.c.L.Lock()
	defer h.c.L.Unlock()
	for h.b.Len() == 0 {
		h.c.Wait()
	}
	// There is definitely something in there
	x := h.b.Front()
	h.b.Remove(x)
	return x.Value
}

func (h *Heap) Push(x any) {
	defer h.c.Signal() // will wake up a popper
	h.c.L.Lock()
	defer h.c.L.Unlock()
	// Add and sort to maintain priority (not really how the heap works)
	h.b.PushBack(x)
}
