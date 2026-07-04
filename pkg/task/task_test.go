package task_test

import (
	"testing"

	"github.com/krwg/gosched/pkg/task"
)

func TestValidate(t *testing.T) {
	ok := &task.Task{
		ID: "a", Name: "A", Duration: 10, Deadline: 100,
		Priority: 5, ArrivalTime: 0,
	}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}

	bad := &task.Task{ID: "", Name: "x", Duration: 1, Deadline: 2, Priority: 1}
	if err := bad.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLaxity(t *testing.T) {
	tk := &task.Task{Duration: 20, Deadline: 100, Remaining: 20}
	if tk.Laxity(50) != 30 {
		t.Fatalf("laxity=%d", tk.Laxity(50))
	}
}

func TestFeasibleAt(t *testing.T) {
	tk := &task.Task{Duration: 20, Deadline: 100, Remaining: 20}
	if !tk.FeasibleAt(70) {
		t.Fatal("should be feasible")
	}
	if tk.FeasibleAt(85) {
		t.Fatal("should miss deadline")
	}
}
