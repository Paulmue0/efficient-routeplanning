package collection

import (
	"container/heap"
)

// This impl. is adapted by me with generics from the go standard library code for heaps.
// https://pkg.go.dev/container/heap#example-package-IntHeap

type Item[T comparable] struct {
	value    T
	priority float64
	index    int
}

func NewItem[T comparable](value T, prio float64, index int) *Item[T] {
	return &Item[T]{value, prio, index}
}

// Implements heap.Interface and holds Items.
type PriorityQueue[T comparable] struct {
	items []*Item[T]
	index map[T]*Item[T]
}

func NewPriorityQueue[T comparable]() *PriorityQueue[T] {
	return &PriorityQueue[T]{
		items: []*Item[T]{},
		index: make(map[T]*Item[T]),
	}
}

type Entry[T comparable] struct {
	Value    T
	Priority float64
}

func (pq *PriorityQueue[T]) Items() []Entry[T] {
	result := make([]Entry[T], len(pq.items))
	for i, item := range pq.items {
		result[i] = Entry[T]{Value: item.value, Priority: item.priority}
	}
	return result
}

func (pq *PriorityQueue[T]) GetValue(item *Item[T]) T {
	return item.value
}

func (pq *PriorityQueue[T]) Len() int {
	return len(pq.items)
}

func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return pq.items[i].priority < pq.items[j].priority
}

func (pq *PriorityQueue[T]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

func (pq *PriorityQueue[T]) Push(x any) {
	item := x.(*Item[T])
	item.index = len(pq.items)
	pq.items = append(pq.items, item)
	pq.index[item.value] = item
}

func (pq *PriorityQueue[T]) PushWithPriority(value T, priority float64) {
	item := &Item[T]{value: value, priority: priority}
	heap.Push(pq, item)
}

func (pq *PriorityQueue[T]) Pop() any {
	n := len(pq.items)
	item := pq.items[n-1]
	pq.items[n-1] = nil
	pq.items = pq.items[:n-1]
	item.index = -1
	delete(pq.index, item.value)
	return item
}

func (pq *PriorityQueue[T]) Update(item *Item[T], priority float64) {
	item.priority = priority
	heap.Fix(pq, item.index)
}

func (pq *PriorityQueue[T]) UpdatePriority(value T, priority float64) {
	if item, ok := pq.index[value]; ok {
		item.priority = priority
		heap.Fix(pq, item.index)
	} else {
		pq.PushWithPriority(value, priority)
	}
}
