package main

import (
	"context"
	"fmt"
	"time"

	"github.com/krwg/gosched/internal/engine"
	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
)

func main() {
	eng := engine.New(engine.Config{Algorithm: scheduler.LLF, WorkerCount: 4})
	for i := 0; i < 5; i++ {
		eng.Tasks() <- &task.Task{
			ID:          fmt.Sprintf("job-%d", i),
			Name:        "Job",
			Duration:    15,
			Deadline:    int64(100 + i*20),
			Priority:    3,
			ArrivalTime: int64(i * 3),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := eng.Run(ctx, nil)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("completed=%d missed=%d\n", res.Metrics.Completed, res.Metrics.MissedDeadlines)
}
