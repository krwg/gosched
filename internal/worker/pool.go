package worker

import (
	"context"
	"sync"

	"github.com/krwg/gosched/pkg/task"
)

type Job struct {
	Task      *task.Task
	StartTime int64
	WorkerID  int
}

type Result struct {
	Task     *task.Task
	WorkerID int
	Missed   bool
	WaitMs   int64
	BusyMs   int64
}

type Pool struct {
	size    int
	jobs    chan Job
	results chan Result
	wg      sync.WaitGroup
}

func NewPool(size int, queueSize int) *Pool {
	if size <= 0 {
		size = 1
	}
	if queueSize <= 0 {
		queueSize = size * 4
	}
	return &Pool{
		size:    size,
		jobs:    make(chan Job, queueSize),
		results: make(chan Result, queueSize),
	}
}

func (p *Pool) Results() <-chan Result { return p.results }

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i+1)
	}
}

func (p *Pool) Submit(ctx context.Context, job Job) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.jobs <- job:
		return nil
	}
}

func (p *Pool) Close() {
	close(p.jobs)
	p.wg.Wait()
	close(p.results)
}

func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			t := job.Task
			wait := job.StartTime - t.ArrivalTime
			if wait < 0 {
				wait = 0
			}

			missed := !t.FeasibleAt(job.StartTime)
			busy := int64(0)
			if !missed {
				t.Status = task.StatusRunning
				t.StartTime = job.StartTime
				t.AssignedWorker = id
				busy = t.Remaining
				t.CompletionTime = job.StartTime + busy
				t.Status = task.StatusCompleted
			} else {
				t.Status = task.StatusMissedDeadline
				t.StartTime = job.StartTime
				t.CompletionTime = job.StartTime
				t.AssignedWorker = id
			}

			select {
			case <-ctx.Done():
				return
			case p.results <- Result{
				Task:     t,
				WorkerID: id,
				Missed:   missed,
				WaitMs:   wait,
				BusyMs:   busy,
			}:
			}
		}
	}
}

func (p *Pool) Size() int { return p.size }
