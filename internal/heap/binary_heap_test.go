package heap_test

import (
	"testing"

	"github.com/krwg/gosched/internal/heap"
)

func TestBinaryHeapPushPop(t *testing.T) {
	h := heap.New(func(a, b int) bool { return a < b })
	for _, v := range []int{5, 3, 8, 1, 9, 2} {
		h.Push(v)
		if !h.Verify() {
			t.Fatalf("heap invariant broken after push %d", v)
		}
	}
	prev := -1
	for !h.Empty() {
		v, ok := h.Pop()
		if !ok {
			t.Fatal("expected pop")
		}
		if v < prev {
			t.Fatalf("out of order: %d after %d", v, prev)
		}
		prev = v
	}
}

func TestBinaryHeapHeapify(t *testing.T) {
	h := heap.New(func(a, b int) bool { return a < b })
	h.Heapify([]int{9, 4, 7, 1, 3, 2, 8})
	if !h.Verify() {
		t.Fatal("heapify violated invariant")
	}
	if v, _ := h.Peek(); v != 1 {
		t.Fatalf("peek=%d want 1", v)
	}
}

func TestBinaryHeapReplaceRoot(t *testing.T) {
	h := heap.New(func(a, b int) bool { return a < b })
	h.Heapify([]int{1, 2, 3, 4})
	h.ReplaceRoot(0)
	if v, _ := h.Peek(); v != 0 {
		t.Fatalf("peek=%d want 0", v)
	}
}

func BenchmarkHeapPushPop(b *testing.B) {
	h := heap.New(func(a, b int) bool { return a < b })
	vals := make([]int, b.N)
	for i := range vals {
		vals[i] = i % 1000
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Push(vals[i])
		_, _ = h.Pop()
	}
}
