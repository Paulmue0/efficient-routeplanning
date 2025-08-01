// Package collection provides a generic priority queue implementation.
package collection

import (
	"container/heap"
)

// This implementation is adapted by me with generics from the Go standard library's
// heap examples.
// https://pkg.go.dev/container/heap#example-package-IntHeap

// Item represents a single item in the priority queue. It holds a value,
// a priority, and its index in the heap. The index is used by the
// heap.Interface implementation to maintain the heap property efficiently.
type Item[T comparable] struct {
	value    T       // The value of the item.
	priority float64 // The priority of the item in the queue.
	index    int     // The index of the item in the heap.
}

// NewItem creates and returns a new Item. This function is typically used
// internally by the PriorityQueue, but can be used to create items directly.
func NewItem[T comparable](value T, prio float64, index int) *Item[T] {
	return &Item[T]{value, prio, index}
}

// PriorityQueue represents a priority queue. It is implemented as a min-heap
// over a slice of Items. It also maintains a map to allow for efficient
// updates of item priorities. The zero value for a PriorityQueue is not ready
// for use; one should be created with NewPriorityQueue.
type PriorityQueue[T comparable] struct {
	items []*Item[T]     // The items in the priority queue, maintained as a heap.
	index map[T]*Item[T] // A map to access items by their value for quick updates.
}

// NewPriorityQueue creates and initializes a new, empty PriorityQueue.
func NewPriorityQueue[T comparable]() *PriorityQueue[T] {
	return &PriorityQueue[T]{
		items: []*Item[T]{},
		index: make(map[T]*Item[T]),
	}
}

// Entry represents the public-facing view of an item in the priority queue,
// containing only the value and its priority.
type Entry[T comparable] struct {
	Value    T
	Priority float64
}

// Items returns a slice of all entries currently in the priority queue.
// The order of the entries is not guaranteed to be sorted by priority.
func (pq *PriorityQueue[T]) Items() []Entry[T] {
	result := make([]Entry[T], len(pq.items))
	for i, item := range pq.items {
		result[i] = Entry[T]{Value: item.value, Priority: item.priority}
	}
	return result
}

// GetValue returns the value of a given Item.
func (pq *PriorityQueue[T]) GetValue(item *Item[T]) T {
	return item.value
}

// GetValue returns the value of a given Item.
func (pq *PriorityQueue[T]) GetPriority(item *Item[T]) float64 {
	return item.priority
}

// Len returns the number of items in the priority queue.
// It is part of the heap.Interface implementation.
func (pq *PriorityQueue[T]) Len() int {
	return len(pq.items)
}

// Less reports whether the item at index i should sort before the item at index j.
// It is part of the heap.Interface implementation and establishes the min-heap property.
func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return pq.items[i].priority < pq.items[j].priority
}

// Swap swaps the items at indices i and j.
// It is part of the heap.Interface implementation.
func (pq *PriorityQueue[T]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

// Push adds an item to the priority queue.
// It is part of the heap.Interface implementation. Users should typically use
// PushWithPriority or UpdatePriority instead of calling this directly.
func (pq *PriorityQueue[T]) Push(x any) {
	item := x.(*Item[T])
	item.index = len(pq.items)
	pq.items = append(pq.items, item)
	pq.index[item.value] = item
}

// PushWithPriority adds a new value with a given priority to the queue.
// It is a convenience method that wraps the standard heap.Push operation.
func (pq *PriorityQueue[T]) PushWithPriority(value T, priority float64) {
	item := &Item[T]{value: value, priority: priority}
	heap.Push(pq, item)
}

// Pop removes and returns the item with the lowest priority from the queue.
// It is part of the heap.Interface implementation. The return value is an any
// type and needs to be type-asserted to *Item[T].
func (pq *PriorityQueue[T]) Pop() any {
	n := len(pq.items)
	item := pq.items[n-1]
	pq.items[n-1] = nil // avoid memory leak
	pq.items = pq.items[:n-1]
	item.index = -1 // for safety
	delete(pq.index, item.value)
	return item
}

// Update modifies the priority of an item in the queue. After changing the
// priority, it re-establishes the heap property.
func (pq *PriorityQueue[T]) Update(item *Item[T], priority float64) {
	item.priority = priority
	heap.Fix(pq, item.index)
}

// UpdatePriority updates the priority of an item identified by its value.
// If the item does not exist in the queue, it is added with the specified priority.
func (pq *PriorityQueue[T]) UpdatePriority(value T, priority float64) {
	if item, ok := pq.index[value]; ok {
		item.priority = priority
		heap.Fix(pq, item.index)
	} else {
		pq.PushWithPriority(value, priority)
	}
}
