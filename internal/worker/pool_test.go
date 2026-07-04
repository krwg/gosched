package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/krwg/gosched/internal/worker"
	"github.com/krwg/gosched/pkg/task"
)

func TestPoolExecute(t *testing.T) {
	pool := worker.NewPool(2, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	go func() {
		tk := &task.Task{
			ID: "j", Name: "J", Duration: 10, Deadline: 100,
			Priority: 5, ArrivalTime: 0, Remaining: 10,
		}
		tk.Normalize()
		_ = pool.Submit(ctx, worker.Job{Task: tk, StartTime: 0, WorkerID: 1})
		pool.Close()
	}()

	select {
	case res := <-pool.Results():
		if res.Missed {
			t.Fatal("should complete")
		}
		if res.Task.Status != task.StatusCompleted {
			t.Fatalf("status=%s", res.Task.Status)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
