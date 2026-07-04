package heap

type LessFunc[T any] func(a, b T) bool

type BinaryHeap[T any] struct {
	data []T
	less LessFunc[T]
}

func New[T any](less LessFunc[T]) *BinaryHeap[T] {
	return &BinaryHeap[T]{less: less}
}

func (h *BinaryHeap[T]) Len() int { return len(h.data) }

func (h *BinaryHeap[T]) Empty() bool { return len(h.data) == 0 }

func (h *BinaryHeap[T]) Peek() (T, bool) {
	var zero T
	if len(h.data) == 0 {
		return zero, false
	}
	return h.data[0], true
}

func (h *BinaryHeap[T]) Push(v T) {
	h.data = append(h.data, v)
	h.siftUp(len(h.data) - 1)
}

func (h *BinaryHeap[T]) Pop() (T, bool) {
	var zero T
	if len(h.data) == 0 {
		return zero, false
	}
	min := h.data[0]
	last := len(h.data) - 1
	h.data[0] = h.data[last]
	h.data = h.data[:last]
	if len(h.data) > 0 {
		h.siftDown(0)
	}
	return min, true
}

func (h *BinaryHeap[T]) ReplaceRoot(v T) {
	if len(h.data) == 0 {
		h.Push(v)
		return
	}
	h.data[0] = v
	h.siftDown(0)
}

func (h *BinaryHeap[T]) Data() []T {
	out := make([]T, len(h.data))
	copy(out, h.data)
	return out
}

func (h *BinaryHeap[T]) Heapify(items []T) {
	h.data = append(h.data[:0], items...)
	for i := parent(len(h.data) - 1); i >= 0; i-- {
		h.siftDown(i)
	}
}

func (h *BinaryHeap[T]) siftUp(i int) {
	for i > 0 {
		p := parent(i)
		if !h.less(h.data[i], h.data[p]) {
			break
		}
		h.data[i], h.data[p] = h.data[p], h.data[i]
		i = p
	}
}

func (h *BinaryHeap[T]) siftDown(i int) {
	n := len(h.data)
	for {
		left := leftChild(i)
		if left >= n {
			break
		}
		smallest := left
		right := rightChild(i)
		if right < n && h.less(h.data[right], h.data[smallest]) {
			smallest = right
		}
		if !h.less(h.data[smallest], h.data[i]) {
			break
		}
		h.data[i], h.data[smallest] = h.data[smallest], h.data[i]
		i = smallest
	}
}

func parent(i int) int   { return (i - 1) / 2 }
func leftChild(i int) int  { return 2*i + 1 }
func rightChild(i int) int { return 2*i + 2 }

func (h *BinaryHeap[T]) Verify() bool {
	for i := range h.data {
		l, r := leftChild(i), rightChild(i)
		if l < len(h.data) && h.less(h.data[l], h.data[i]) {
			return false
		}
		if r < len(h.data) && h.less(h.data[r], h.data[i]) {
			return false
		}
	}
	return true
}
