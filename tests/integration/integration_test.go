package integration_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/krwg/gosched/internal/engine"
	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
	"github.com/krwg/gosched/pkg/visualizer"
)

func fixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "fixtures", name)
}

func TestFixtureSchedule(t *testing.T) {
	tasks, err := task.LoadTasksFromFile(fixturePath("tasks.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, algo := range []scheduler.Algorithm{scheduler.RM, scheduler.EDF, scheduler.LLF} {
		res, err := engine.RunQuick(context.Background(), engine.Config{
			Algorithm: algo, WorkerCount: 4,
		}, tasks)
		if err != nil {
			t.Fatalf("%s: %v", algo, err)
		}
		if res.Metrics.TasksProcessed != len(tasks) {
			t.Fatalf("%s processed=%d", algo, res.Metrics.TasksProcessed)
		}
		ascii := visualizer.RenderASCII(res.Gantt, res.Metrics.TotalSimulatedMs, 32)
		if len(ascii) < 20 {
			t.Fatalf("short ascii output")
		}
	}
}

func TestHighLoadEDF(t *testing.T) {
	tasks := make([]*task.Task, 0, 100)
	for i := 0; i < 100; i++ {
		tasks = append(tasks, &task.Task{
			ID:          "t",
			Name:        "T",
			Duration:    5,
			Deadline:    int64(50 + i*2),
			Priority:    1 + i%10,
			ArrivalTime: int64(i % 20),
		})
	}
	res, err := engine.RunQuick(context.Background(), engine.Config{
		Algorithm: scheduler.EDF, WorkerCount: 4,
	}, tasks)
	if err != nil {
		t.Fatal(err)
	}
	if res.Metrics.MaxQueueDepth == 0 {
		t.Fatal("expected queue depth > 0")
	}
}
