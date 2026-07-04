package scheduler

import (
	"fmt"
	"strings"

	"github.com/krwg/gosched/pkg/task"
)

type Algorithm string

const (
	RM  Algorithm = "rm"
	EDF Algorithm = "edf"
	LLF Algorithm = "llf"
)

func ParseAlgorithm(s string) (Algorithm, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "rm", "rate-monotonic":
		return RM, nil
	case "edf", "earliest-deadline-first":
		return EDF, nil
	case "llf", "least-laxity-first":
		return LLF, nil
	default:
		return "", fmt.Errorf("unknown algorithm %q (use rm, edf, or llf)", s)
	}
}

type Scheduler interface {
	Name() string
	Algorithm() Algorithm
	Push(t *task.Task)
	Pop() (*task.Task, bool)
	Peek() (*task.Task, bool)
	Len() int
	OnTick(currentTime int64)
}

func New(algo Algorithm) Scheduler {
	switch algo {
	case RM:
		return newRM()
	case EDF:
		return newEDF()
	case LLF:
		return newLLF()
	default:
		return newEDF()
	}
}

func BruteforceEDFOrder(tasks []*task.Task) []*task.Task {
	if len(tasks) == 0 {
		return nil
	}
	sorted := make([]*task.Task, len(tasks))
	copy(sorted, tasks)
	for i := 1; i < len(sorted); i++ {
		j := i
		for j > 0 && lessEDF(sorted[j], sorted[j-1]) {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			j--
		}
	}
	return sorted
}

func lessEDF(a, b *task.Task) bool {
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
}

func PopSequence(s Scheduler) []*task.Task {
	var out []*task.Task
	for {
		t, ok := s.Pop()
		if !ok {
			break
		}
		out = append(out, t)
	}
	return out
}
