package datastruct

// NewHeap creates a new Heap instance that is configured to use the
// specified less function to order items.
func NewHeap[T any](less LessFunc[T]) *Heap[T] {
	return &Heap[T]{
		less: less,
	}
}

// Heap is a datastructure that orders items when inserted.
type Heap[T any] struct {
	less  LessFunc[T]
	items []T
}

// IsEmpty returns true if there are no items in this Heap.
func (h *Heap[T]) IsEmpty() bool {
	return len(h.items) == 0
}

// Clear removes all items from this Heap.
func (h *Heap[T]) Clear() {
	h.items = h.items[:0]
}

// Push adds a new item to this Heap.
func (h *Heap[T]) Push(value T) {
	h.items = append(h.items, value)
	h.siftUp(value, len(h.items)-1)
}

// Pop removes the smallest item from this Heap and returns it.
func (h *Heap[T]) Pop() T {
	result := h.items[0]
	if len(h.items) > 1 {
		h.items[0] = h.items[len(h.items)-1]
		h.items = h.items[:len(h.items)-1]
		h.siftDown(h.items[0], 0)
	} else {
		h.items = h.items[:len(h.items)-1]
	}
	return result
}

// Peek returns the samllest item from this Heap without actually
// removing it.
func (h *Heap[T]) Peek() T {
	return h.items[0]
}

func (h *Heap[T]) siftUp(value T, index int) {
	for index > 0 {
		parentIndex := (index - 1) / 2
		parentValue := h.items[parentIndex]
		if !h.less(value, parentValue) {
			return
		}
		h.items[index] = parentValue
		h.items[parentIndex] = value
		index = parentIndex
	}
}

func (h *Heap[T]) siftDown(value T, index int) {
	leftChildIndex := index*2 + 1
	for leftChildIndex < len(h.items) {
		smallestIndex := index

		leftValue := h.items[leftChildIndex]
		if h.less(leftValue, value) {
			smallestIndex = leftChildIndex
		}

		rightChildIndex := leftChildIndex + 1
		if rightChildIndex < len(h.items) {
			rightValue := h.items[rightChildIndex]
			if h.less(rightValue, value) && h.less(rightValue, leftValue) {
				smallestIndex = rightChildIndex
			}
		}

		if smallestIndex == index {
			return
		}

		h.items[index] = h.items[smallestIndex]
		h.items[smallestIndex] = value

		index = smallestIndex
		leftChildIndex = index*2 + 1
	}
}
