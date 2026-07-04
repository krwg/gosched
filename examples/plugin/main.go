package main

import (
	"github.com/krwg/gosched/internal/plugin"
	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
)

var Scheduler plugin.Factory = deadlinePriorityPlugin{}

type deadlinePriorityPlugin struct{}

func (deadlinePriorityPlugin) Name() string { return "Deadline Priority Plugin" }

func (deadlinePriorityPlugin) Build() scheduler.Scheduler {
	return scheduler.NewCustom("deadline-priority", func(a, b *task.Task) bool {
		if a.Deadline != b.Deadline {
			return a.Deadline < b.Deadline
		}
		return a.Priority < b.Priority
	})
}
