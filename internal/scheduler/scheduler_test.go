package scheduler_test

import (
	"testing"

	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
)

func sampleTasks() []*task.Task {
	return []*task.Task{
		{ID: "1", Name: "A", Duration: 10, Deadline: 100, Priority: 2, ArrivalTime: 0, Period: 100},
		{ID: "2", Name: "B", Duration: 10, Deadline: 50, Priority: 3, ArrivalTime: 0, Period: 50},
		{ID: "3", Name: "C", Duration: 10, Deadline: 50, Priority: 1, ArrivalTime: 0, Period: 50},
	}
}

func TestEDFOrdering(t *testing.T) {
	s := scheduler.New(scheduler.EDF)
	for _, tk := range sampleTasks() {
		tk.Normalize()
		s.Push(tk)
	}
	got := scheduler.PopSequence(s)
	if len(got) != 3 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].ID != "2" && got[0].ID != "3" {
		t.Fatalf("earliest deadline first, got %s", got[0].ID)
	}
	if got[0].Deadline != 50 {
		t.Fatalf("deadline=%d", got[0].Deadline)
	}

	want := scheduler.BruteforceEDFOrder(sampleTasks())
	if len(want) != len(got) {
		t.Fatal("length mismatch")
	}
	for i := range want {
		if want[i].ID != got[i].ID {
			t.Fatalf("edf[%d]=%s bruteforce=%s", i, got[i].ID, want[i].ID)
		}
	}
}

func TestRMShorterPeriodFirst(t *testing.T) {
	s := scheduler.New(scheduler.RM)
	for _, tk := range sampleTasks() {
		tk.Normalize()
		s.Push(tk)
	}
	got := scheduler.PopSequence(s)
	if got[0].Period > got[1].Period {
		t.Fatalf("rm order wrong: %d then %d", got[0].Period, got[1].Period)
	}
}

func TestLLFTieBreak(t *testing.T) {
	s := scheduler.New(scheduler.LLF)
	tasks := []*task.Task{
		{ID: "a", Name: "A", Duration: 30, Deadline: 100, ArrivalTime: 0, Remaining: 30},
		{ID: "b", Name: "B", Duration: 20, Deadline: 80, ArrivalTime: 0, Remaining: 20},
	}
	for _, tk := range tasks {
		tk.Normalize()
		s.Push(tk)
	}
	s.OnTick(0)
	got := scheduler.PopSequence(s)
	if got[0].ID != "b" {
		t.Fatalf("least laxity expected b, got %s", got[0].ID)
	}
}

func TestSameDeadlineEDF(t *testing.T) {
	s := scheduler.New(scheduler.EDF)
	for _, id := range []string{"z", "a", "m"} {
		s.Push(&task.Task{ID: id, Name: id, Duration: 5, Deadline: 40, Priority: 5, ArrivalTime: 0})
	}
	got := scheduler.PopSequence(s)
	if got[0].ID != "a" {
		t.Fatalf("tie-break by id: %s", got[0].ID)
	}
}

func TestParseAlgorithm(t *testing.T) {
	if _, err := scheduler.ParseAlgorithm("edf"); err != nil {
		t.Fatal(err)
	}
	if _, err := scheduler.ParseAlgorithm("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func BenchmarkEDFPushPop(b *testing.B) {
	s := scheduler.New(scheduler.EDF)
	tasks := make([]*task.Task, 1000)
	for i := range tasks {
		tasks[i] = &task.Task{
			ID: "t", Duration: 5, Deadline: int64(100 + i), ArrivalTime: 0, Priority: 5,
		}
		tasks[i].Normalize()
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, tk := range tasks {
			s.Push(tk)
		}
		for s.Len() > 0 {
			_, _ = s.Pop()
		}
	}
}

func BenchmarkRMPushPop(b *testing.B) {
	s := scheduler.New(scheduler.RM)
	tasks := make([]*task.Task, 1000)
	for i := range tasks {
		tasks[i] = &task.Task{
			ID: "t", Duration: 5, Deadline: int64(100 + i), Period: int64(50 + i%30),
			ArrivalTime: 0, Priority: 1 + i%10,
		}
		tasks[i].Normalize()
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, tk := range tasks {
			s.Push(tk)
		}
		for s.Len() > 0 {
			_, _ = s.Pop()
		}
	}
}
