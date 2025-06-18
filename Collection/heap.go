package collection

import (
	"container/heap"
)

// This impl. is adapted by me with generics from the go standard library code for heaps.
// https://pkg.go.dev/container/heap#example-package-IntHeap

// An Item is something we manage in a priority queue.
type Item[T comparable] struct {
	Value    T
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

func (pq *PriorityQueue[T]) Len() int {
	return len(pq.items)
}

func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return pq.items[i].priority > pq.items[j].priority
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
	pq.index[item.Value] = item
}

func (pq *PriorityQueue[T]) PushWithPriority(value T, priority float64) {
	item := &Item[T]{Value: value, priority: priority}
	heap.Push(pq, item)
}

func (pq *PriorityQueue[T]) Pop() any {
	n := len(pq.items)
	item := pq.items[n-1]
	pq.items[n-1] = nil
	pq.items = pq.items[:n-1]
	item.index = -1
	delete(pq.index, item.Value)
	return item
}

func (pq *PriorityQueue[T]) Update(item *Item[T], value T, priority float64) {
	item.Value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

func (pq *PriorityQueue[T]) UpdatePriority(value T, priority float64) {
	if item, ok := pq.index[value]; ok {
		item.priority = priority
		heap.Fix(pq, item.index)
	}
}
