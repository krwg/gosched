package main

import (
	"context"
	"fmt"

	"github.com/krwg/gosched/internal/engine"
	"github.com/krwg/gosched/internal/metrics"
	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
)

func main() {
	tasks := []*task.Task{
		{ID: "render", Name: "Render", Duration: 40, Deadline: 120, Priority: 1, ArrivalTime: 0},
		{ID: "sync", Name: "Sync", Duration: 20, Deadline: 80, Priority: 2, ArrivalTime: 5},
	}
	res, err := engine.RunQuick(context.Background(), engine.Config{
		Algorithm: scheduler.EDF, WorkerCount: 2,
	}, tasks)
	if err != nil {
		panic(err)
	}
	fmt.Print(metrics.FormatReport(res.Metrics))
}
