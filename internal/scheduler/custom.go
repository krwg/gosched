package scheduler

import (
	"github.com/krwg/gosched/internal/heap"
	"github.com/krwg/gosched/pkg/task"
)

type customScheduler struct {
	name string
	algo Algorithm
	h    *heap.BinaryHeap[*task.Task]
}

func NewCustom(name string, less func(a, b *task.Task) bool) Scheduler {
	return &customScheduler{
		name: name,
		algo: Algorithm("custom"),
		h:    heap.New(less),
	}
}

func (s *customScheduler) Name() string            { return s.name }
func (s *customScheduler) Algorithm() Algorithm    { return s.algo }
func (s *customScheduler) Push(t *task.Task)       { s.h.Push(t.Clone()) }
func (s *customScheduler) Pop() (*task.Task, bool) { return s.h.Pop() }
func (s *customScheduler) Peek() (*task.Task, bool) { return s.h.Peek() }
func (s *customScheduler) Len() int                { return s.h.Len() }
func (s *customScheduler) OnTick(int64)            {}
