package main

import (
	"context"
	"fmt"

	"github.com/krwg/gosched/internal/engine"
	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
)

func main() {
	tasks := []*task.Task{
		{ID: "1", Name: "Sensor", Duration: 25, Deadline: 100, Priority: 1, ArrivalTime: 0, Period: 100},
		{ID: "2", Name: "Actuator", Duration: 15, Deadline: 60, Priority: 2, ArrivalTime: 0, Period: 60},
		{ID: "3", Name: "Logger", Duration: 10, Deadline: 60, Priority: 3, ArrivalTime: 0, Period: 60},
	}
	ctx := context.Background()
	for _, algo := range []scheduler.Algorithm{scheduler.RM, scheduler.EDF, scheduler.LLF} {
		res, err := engine.RunQuick(ctx, engine.Config{Algorithm: algo, WorkerCount: 2}, tasks)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: completed=%d missed=%d wait=%.0fms depth=%d\n",
			algo, res.Metrics.Completed, res.Metrics.MissedDeadlines,
			res.Metrics.AverageWaitMs, res.Metrics.MaxQueueDepth)
	}
}
