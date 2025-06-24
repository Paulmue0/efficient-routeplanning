package collection

import (
	"container/heap"
	"testing"
)

func TestPriorityQueue_BasicOperations(t *testing.T) {
	pq := NewPriorityQueue[string]()
	heap.Init(pq)

	items := []*Item[string]{
		{value: "low", priority: 1.0},
		{value: "medium", priority: 5.0},
		{value: "high", priority: 10.0},
	}

	for _, item := range items {
		heap.Push(pq, item)
	}

	expectedOrder := []string{"low", "medium", "high"}
	for _, expected := range expectedOrder {
		item := heap.Pop(pq).(*Item[string])
		if item.value != expected {
			t.Errorf("expected %s, got %s", expected, item.value)
		}
	}
}

func TestPriorityQueue_UpdatePriority(t *testing.T) {
	pq := NewPriorityQueue[string]()
	heap.Init(pq)

	low := &Item[string]{value: "task", priority: 1.0}
	heap.Push(pq, low)

	pq.Update(low, 100.0)

	if item, ok := pq.index["task"]; !ok || item != low {
		t.Errorf("index map not updated correctly for 'task'")
	}

	item := heap.Pop(pq).(*Item[string])
	if item.value != "task" || item.priority != 100.0 {
		t.Errorf("unexpected update result: %+v", item)
	}
}

func TestPriorityQueue_EqualPriorities(t *testing.T) {
	pq := NewPriorityQueue[string]()
	heap.Init(pq)

	item1 := &Item[string]{value: "first", priority: 10.0}
	item2 := &Item[string]{value: "second", priority: 10.0}

	heap.Push(pq, item1)
	heap.Push(pq, item2)

	values := []string{
		heap.Pop(pq).(*Item[string]).value,
		heap.Pop(pq).(*Item[string]).value,
	}

	if (values[0] != "first" && values[0] != "second") ||
		(values[1] != "first" && values[1] != "second") ||
		values[0] == values[1] {
		t.Errorf("unexpected order for equal priorities: %v", values)
	}
}

func TestPriorityQueue_EmptyPop(t *testing.T) {
	pq := NewPriorityQueue[string]()
	heap.Init(pq)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on Pop from empty queue, got none")
		}
	}()
	_ = heap.Pop(pq)
}

func TestPriorityQueue_IndexTracking(t *testing.T) {
	pq := NewPriorityQueue[string]()
	heap.Init(pq)

	item := &Item[string]{value: "test", priority: 1.0}
	heap.Push(pq, item)

	if item.index != 0 {
		t.Errorf("expected index to be 0, got %d", item.index)
	}

	heap.Pop(pq)

	if item.index != -1 {
		t.Errorf("expected index to be -1 after pop, got %d", item.index)
	}
}
