package plugin_test

import (
	"testing"

	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
)

type stubFactory struct{}

func (stubFactory) Name() string { return "stub" }

func (stubFactory) Build() scheduler.Scheduler {
	return scheduler.NewCustom("stub", func(a, b *task.Task) bool {
		return a.ID < b.ID
	})
}

func TestCustomScheduler(t *testing.T) {
	s := stubFactory{}.Build()
	s.Push(&task.Task{ID: "b", Name: "B", Duration: 5, Deadline: 50, Priority: 1, ArrivalTime: 0})
	s.Push(&task.Task{ID: "a", Name: "A", Duration: 5, Deadline: 50, Priority: 1, ArrivalTime: 0})
	first, ok := s.Pop()
	if !ok || first.ID != "a" {
		t.Fatalf("got %v", first)
	}
}
