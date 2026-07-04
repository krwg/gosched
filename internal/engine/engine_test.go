package engine_test

import (
	"context"
	"testing"

	"github.com/krwg/gosched/internal/engine"
	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
)

func TestRunQuickEDF(t *testing.T) {
	tasks := []*task.Task{
		{ID: "task-1", Name: "Rendering", Duration: 100, Deadline: 500, Priority: 1, ArrivalTime: 0},
		{ID: "task-2", Name: "Data Backup", Duration: 50, Deadline: 200, Priority: 3, ArrivalTime: 10},
		{ID: "task-3", Name: "Log Rotation", Duration: 30, Deadline: 150, Priority: 2, ArrivalTime: 5},
	}
	res, err := engine.RunQuick(context.Background(), engine.Config{
		Algorithm: scheduler.EDF, WorkerCount: 4,
	}, tasks)
	if err != nil {
		t.Fatal(err)
	}
	if res.Metrics.TasksProcessed != 3 {
		t.Fatalf("processed=%d", res.Metrics.TasksProcessed)
	}
	if len(res.Gantt) == 0 {
		t.Fatal("expected gantt segments")
	}
}

func TestMissedDeadline(t *testing.T) {
	tasks := []*task.Task{
		{ID: "a", Name: "A", Duration: 80, Deadline: 100, Priority: 1, ArrivalTime: 0},
		{ID: "b", Name: "B", Duration: 80, Deadline: 100, Priority: 2, ArrivalTime: 0},
	}
	res, err := engine.RunQuick(context.Background(), engine.Config{
		Algorithm: scheduler.EDF, WorkerCount: 1,
	}, tasks)
	if err != nil {
		t.Fatal(err)
	}
	if res.Metrics.MissedDeadlines == 0 {
		t.Fatalf("missed=%d", res.Metrics.MissedDeadlines)
	}
}

func TestEmptyQueue(t *testing.T) {
	res, err := engine.RunQuick(context.Background(), engine.Config{
		Algorithm: scheduler.EDF, WorkerCount: 2,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Metrics.TasksProcessed != 0 {
		t.Fatalf("processed=%d", res.Metrics.TasksProcessed)
	}
}

func TestAsyncSubmit(t *testing.T) {
	eng := engine.New(engine.Config{Algorithm: scheduler.EDF, WorkerCount: 2})
	eng.Tasks() <- &task.Task{
		ID: "x", Name: "X", Duration: 10, Deadline: 100, Priority: 5, ArrivalTime: 0,
	}
	res, err := eng.Run(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Metrics.TasksProcessed != 1 {
		t.Fatalf("processed=%d", res.Metrics.TasksProcessed)
	}
}

func TestAllAlgorithms(t *testing.T) {
	tasks := []*task.Task{
		{ID: "a", Name: "A", Duration: 20, Deadline: 80, Priority: 2, ArrivalTime: 0},
		{ID: "b", Name: "B", Duration: 15, Deadline: 60, Priority: 4, ArrivalTime: 5},
	}
	for _, algo := range []scheduler.Algorithm{scheduler.RM, scheduler.EDF, scheduler.LLF} {
		res, err := engine.RunQuick(context.Background(), engine.Config{Algorithm: algo, WorkerCount: 2}, tasks)
		if err != nil {
			t.Fatalf("%s: %v", algo, err)
		}
		if res.Metrics.TasksProcessed != 2 {
			t.Fatalf("%s processed=%d", algo, res.Metrics.TasksProcessed)
		}
	}
}

func TestResultJSONRoundTrip(t *testing.T) {
	tasks := []*task.Task{
		{ID: "a", Name: "A", Duration: 10, Deadline: 50, Priority: 1, ArrivalTime: 0},
	}
	res, err := engine.RunQuick(context.Background(), engine.Config{Algorithm: scheduler.EDF, WorkerCount: 1}, tasks)
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/result.json"
	if err := engine.SaveResultJSON(path, res); err != nil {
		t.Fatal(err)
	}
	loaded, err := engine.LoadResultJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Metrics.Completed != res.Metrics.Completed {
		t.Fatalf("completed=%d", loaded.Metrics.Completed)
	}
}
