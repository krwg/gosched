package scheduler

import (
	"github.com/krwg/gosched/internal/heap"
	"github.com/krwg/gosched/pkg/task"
)

type edfScheduler struct {
	h *heap.BinaryHeap[*task.Task]
}

func newEDF() *edfScheduler {
	return &edfScheduler{
		h: heap.New(func(a, b *task.Task) bool {
			if a.Deadline != b.Deadline {
				return a.Deadline < b.Deadline
			}
			if a.ArrivalTime != b.ArrivalTime {
				return a.ArrivalTime < b.ArrivalTime
			}
			if a.Priority != b.Priority {
				return a.Priority < b.Priority
			}
			return a.ID < b.ID
		}),
	}
}

func (s *edfScheduler) Name() string            { return "Earliest Deadline First (EDF)" }
func (s *edfScheduler) Algorithm() Algorithm    { return EDF }
func (s *edfScheduler) Push(t *task.Task)       { s.h.Push(t.Clone()) }
func (s *edfScheduler) Pop() (*task.Task, bool) { return s.h.Pop() }
func (s *edfScheduler) Peek() (*task.Task, bool) { return s.h.Peek() }
func (s *edfScheduler) Len() int                { return s.h.Len() }
func (s *edfScheduler) OnTick(int64)            {}
