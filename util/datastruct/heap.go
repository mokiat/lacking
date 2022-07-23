package datastruct

func NewHeap[T any](less LessFunc[T]) *Heap[T] {
	return &Heap[T]{
		less: less,
	}
}

type Heap[T any] struct {
	less  LessFunc[T]
	items []T
}

func (h *Heap[T]) Empty() bool {
	return len(h.items) == 0
}

func (h *Heap[T]) Push(value T) {
	h.items = append(h.items, value)
	h.siftUp(value, len(h.items)-1)
}

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
