package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type Status string

const (
	StatusPending        Status = "pending"
	StatusRunning        Status = "running"
	StatusCompleted      Status = "completed"
	StatusMissedDeadline Status = "missed_deadline"
)

type Task struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Duration    int64  `json:"duration"`
	Deadline    int64  `json:"deadline"`
	Priority    int    `json:"priority"`
	ArrivalTime int64  `json:"arrival_time"`
	Period      int64  `json:"period,omitempty"`

	Status         Status `json:"status"`
	Remaining      int64  `json:"remaining,omitempty"`
	WaitTime       int64  `json:"wait_time,omitempty"`
	StartTime      int64  `json:"start_time,omitempty"`
	CompletionTime int64  `json:"completion_time,omitempty"`
	AssignedWorker int    `json:"assigned_worker,omitempty"`
}

func (t *Task) Validate() error {
	if t.ID == "" {
		return errors.New("task id is required")
	}
	if t.Name == "" {
		return errors.New("task name is required")
	}
	if t.Duration <= 0 {
		return fmt.Errorf("task %s: duration must be positive", t.ID)
	}
	if t.Deadline <= 0 {
		return fmt.Errorf("task %s: deadline must be positive", t.ID)
	}
	if t.Priority < 1 || t.Priority > 10 {
		return fmt.Errorf("task %s: priority must be between 1 and 10", t.ID)
	}
	if t.ArrivalTime < 0 {
		return fmt.Errorf("task %s: arrival_time cannot be negative", t.ID)
	}
	if t.ArrivalTime+t.Duration > t.Deadline {
		return fmt.Errorf("task %s: cannot meet deadline even if started at arrival", t.ID)
	}
	return nil
}

func (t *Task) Normalize() {
	if t.Period <= 0 {
		t.Period = t.Deadline - t.ArrivalTime
		if t.Period <= 0 {
			t.Period = t.Duration
		}
	}
	if t.Remaining == 0 {
		t.Remaining = t.Duration
	}
	if t.Status == "" {
		t.Status = StatusPending
	}
}

func (t *Task) Clone() *Task {
	cp := *t
	return &cp
}

func (t *Task) Laxity(currentTime int64) int64 {
	return t.Deadline - currentTime - t.Remaining
}

func (t *Task) FeasibleAt(currentTime int64) bool {
	return currentTime+t.Remaining <= t.Deadline
}

func LoadTasksFromFile(path string) ([]*Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tasks []*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	for _, t := range tasks {
		t.Normalize()
		if err := t.Validate(); err != nil {
			return nil, err
		}
	}
	return tasks, nil
}

func NowMillis() int64 {
	return time.Now().UnixMilli()
}
