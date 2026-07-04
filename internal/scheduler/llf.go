package scheduler

import (
	"github.com/krwg/gosched/internal/heap"
	"github.com/krwg/gosched/pkg/task"
)

type llfScheduler struct {
	h           *heap.BinaryHeap[*task.Task]
	currentTime int64
}

func newLLF() *llfScheduler {
	s := &llfScheduler{}
	s.h = heap.New(func(a, b *task.Task) bool {
		la := a.Laxity(s.currentTime)
		lb := b.Laxity(s.currentTime)
		if la != lb {
			return la < lb
		}
		if a.Deadline != b.Deadline {
			return a.Deadline < b.Deadline
		}
		return a.ID < b.ID
	})
	return s
}

func (s *llfScheduler) Name() string            { return "Least Laxity First (LLF)" }
func (s *llfScheduler) Algorithm() Algorithm    { return LLF }
func (s *llfScheduler) Push(t *task.Task)       { s.h.Push(t.Clone()) }
func (s *llfScheduler) Pop() (*task.Task, bool) { return s.h.Pop() }
func (s *llfScheduler) Peek() (*task.Task, bool) { return s.h.Peek() }
func (s *llfScheduler) Len() int                { return s.h.Len() }

func (s *llfScheduler) OnTick(currentTime int64) {
	if s.h.Len() <= 1 {
		s.currentTime = currentTime
		return
	}
	s.currentTime = currentTime
	items := s.h.Data()
	s.h.Heapify(items)
}
