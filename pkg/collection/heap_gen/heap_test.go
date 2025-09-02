// package collection_test contains the tests for the collection package.
package collection

import (
	"container/heap"
	"fmt"
	"sort"
	"testing"
)

// TestPriorityQueue_BasicOperations verifies the fundamental push and pop
// functionality of the PriorityQueue. It ensures that items are popped in
// ascending order of their priority, confirming the min-heap property.
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

// TestPriorityQueue_UpdatePriority tests the ability to update an item's
// priority and have the queue re-order itself correctly. It also checks
// that the internal index map is kept consistent after the update.
func TestPriorityQueue_UpdatePriority(t *testing.T) {
	pq := NewPriorityQueue[string]()
	heap.Init(pq)

	low := &Item[string]{value: "task", priority: 1.0}
	heap.Push(pq, low)

	// Update the priority to a higher value.
	pq.Update(low, 100.0)

	if item, ok := pq.index["task"]; !ok || item != low {
		t.Errorf("index map not updated correctly for 'task'")
	}

	// Pop the item and check if the priority was updated.
	item := heap.Pop(pq).(*Item[string])
	if item.value != "task" || item.priority != 100.0 {
		t.Errorf("unexpected update result: %+v", item)
	}
}

// TestPriorityQueue_EqualPriorities checks the behavior of the queue when two
// items have the same priority. The heap implementation does not guarantee a
// stable order for items of equal priority, so this test ensures that both
// items are eventually popped without error.
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

	// Check that both "first" and "second" were popped, regardless of order.
	if (values[0] != "first" && values[0] != "second") ||
		(values[1] != "first" && values[1] != "second") ||
		values[0] == values[1] {
		t.Errorf("unexpected order for equal priorities: %v", values)
	}
}

// TestPriorityQueue_EmptyPop verifies that attempting to Pop from an empty
// queue triggers a panic. This is the expected behavior from the underlying
// container/heap package.
func TestPriorityQueue_EmptyPop(t *testing.T) {
	pq := NewPriorityQueue[string]()
	heap.Init(pq)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on Pop from empty queue, got none")
		}
	}()
	// This line should panic.
	_ = heap.Pop(pq)
}

// TestPriorityQueue_IndexTracking ensures that the 'index' field of an Item
// is correctly managed by the PriorityQueue. The index should be set upon
// Push and reset to -1 upon Pop, which is crucial for the heap.Fix operation.
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

// ExamplePriorityQueue demonstrates the basic usage of a PriorityQueue.
// It shows how to push items with different priorities and how they are
// popped in ascending order of priority (min-heap).
func ExamplePriorityQueue() {
	// Create a new priority queue.
	pq := NewPriorityQueue[string]()

	// Push some items onto the queue with different priorities.
	// Note: In a real application, you would likely use pq.PushWithPriority
	// instead of creating Item structs directly.
	heap.Push(pq, &Item[string]{value: "medium", priority: 5.0})
	heap.Push(pq, &Item[string]{value: "high", priority: 10.0})
	heap.Push(pq, &Item[string]{value: "low", priority: 1.0})

	// Pop items from the queue until it is empty.
	// They will be returned in order of priority, from lowest to highest.
	for pq.Len() > 0 {
		item := heap.Pop(pq).(*Item[string])
		fmt.Printf("Value: %s, Priority: %.1f\n", item.value, item.priority)
	}

	// Output:
	// Value: low, Priority: 1.0
	// Value: medium, Priority: 5.0
	// Value: high, Priority: 10.0
}

// ExamplePriorityQueue_UpdatePriority demonstrates how to add an item and
// then update its priority.
func ExamplePriorityQueue_UpdatePriority() {
	pq := NewPriorityQueue[string]()

	// Add two items to the queue.
	pq.PushWithPriority("first", 5.0)
	pq.PushWithPriority("second", 10.0)

	// Update the priority of "first" to be higher than "second".
	pq.UpdatePriority("first", 15.0)

	// Pop the items to see the new order.
	item1 := heap.Pop(pq).(*Item[string])
	item2 := heap.Pop(pq).(*Item[string])

	fmt.Printf("First out: %s\n", item1.value)
	fmt.Printf("Second out: %s\n", item2.value)

	// Output:
	// First out: second
	// Second out: first
}

// ExamplePriorityQueue_equalPriority shows how the queue behaves with items
// of equal priority. The order for items with the same priority is not
// guaranteed. To have a deterministic output for the example, we pop all
// items and sort them before printing.
func ExamplePriorityQueue_equalPriority() {
	pq := NewPriorityQueue[string]()

	pq.PushWithPriority("job_a", 10.0)
	pq.PushWithPriority("job_b", 5.0)
	pq.PushWithPriority("job_c", 10.0)

	var results []string
	for pq.Len() > 0 {
		item := heap.Pop(pq).(*Item[string])
		results = append(results, item.value)
	}

	// The pop order for job_a and job_c isn't guaranteed.
	// For a stable test, we can sort the results.
	sort.Strings(results)
	fmt.Println(results)

	// Output:
	// [job_a job_b job_c]
}
