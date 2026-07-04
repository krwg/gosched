package scheduler

import (
	"github.com/krwg/gosched/internal/heap"
	"github.com/krwg/gosched/pkg/task"
)

type rmScheduler struct {
	h *heap.BinaryHeap[*task.Task]
}

func newRM() *rmScheduler {
	return &rmScheduler{
		h: heap.New(func(a, b *task.Task) bool {
			if a.Period != b.Period {
				return a.Period < b.Period
			}
			if a.Priority != b.Priority {
				return a.Priority < b.Priority
			}
			if a.ArrivalTime != b.ArrivalTime {
				return a.ArrivalTime < b.ArrivalTime
			}
			return a.ID < b.ID
		}),
	}
}

func (s *rmScheduler) Name() string            { return "Rate Monotonic (RM)" }
func (s *rmScheduler) Algorithm() Algorithm    { return RM }
func (s *rmScheduler) Push(t *task.Task)       { s.h.Push(t.Clone()) }
func (s *rmScheduler) Pop() (*task.Task, bool) { return s.h.Pop() }
func (s *rmScheduler) Peek() (*task.Task, bool) { return s.h.Peek() }
func (s *rmScheduler) Len() int                { return s.h.Len() }
func (s *rmScheduler) OnTick(int64)            {}
