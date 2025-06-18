package collection

import (
	"container/heap"
)

// This impl. is adapted by me with generics from the go standard library code for heaps.
// https://pkg.go.dev/container/heap#example-package-IntHeap

// An Item is something we manage in a priority queue.
type Item[T comparable] struct {
	value    T
	priority float64
	index    int
}

// Implements heap.Interface and holds Items.
type PriorityQueue[T comparable] []*Item[T]

func (pq PriorityQueue[T]) Len() int { return len(pq) }

func (pq PriorityQueue[T]) Less(i, j int) bool {
	// As this is a min heap we want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue[T]) Push(x any) {
	n := len(*pq)
	item := x.(*Item[T])
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue[T]) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // don't stop the GC from reclaiming the item eventually
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue[T]) update(item *Item[T], value T, priority float64) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}
