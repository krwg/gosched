package engine

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/krwg/gosched/internal/metrics"
	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
)

type Config struct {
	Algorithm   scheduler.Algorithm
	WorkerCount int
	Custom      scheduler.Scheduler
}

type GanttSegment struct {
	WorkerID  int    `json:"worker_id"`
	TaskID    string `json:"task_id"`
	TaskName  string `json:"task_name"`
	StartMs   int64  `json:"start_ms"`
	EndMs     int64  `json:"end_ms"`
	Algorithm string `json:"algorithm"`
	Missed    bool   `json:"missed"`
}

type Result struct {
	Metrics     metrics.Snapshot `json:"metrics"`
	Gantt       []GanttSegment   `json:"gantt"`
	MissedTasks []*task.Task     `json:"missed_tasks"`
	Completed   []*task.Task     `json:"completed_tasks"`
}

type Engine struct {
	cfg       Config
	sched     scheduler.Scheduler
	collector *metrics.Collector
	taskIn    chan *task.Task
	gantt     []GanttSegment
	missed    []*task.Task
	completed []*task.Task
	mu        sync.Mutex
}

func New(cfg Config) *Engine {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 4
	}
	sched := scheduler.New(cfg.Algorithm)
	if cfg.Custom != nil {
		sched = cfg.Custom
	}
	eng := &Engine{
		cfg:       cfg,
		sched:     sched,
		collector: metrics.New(string(cfg.Algorithm)),
		taskIn:    make(chan *task.Task, 256),
	}
	return eng
}

func (e *Engine) Tasks() chan<- *task.Task { return e.taskIn }

func (e *Engine) Run(ctx context.Context, seed []*task.Task) (*Result, error) {
	e.collector.SetAlgorithm(e.sched.Name())

	arrivals := make([]*task.Task, 0, len(seed))
	for _, t := range seed {
		cp := t.Clone()
		cp.Normalize()
		if err := cp.Validate(); err != nil {
			return nil, err
		}
		arrivals = append(arrivals, cp)
	}
	sort.Slice(arrivals, func(i, j int) bool {
		if arrivals[i].ArrivalTime != arrivals[j].ArrivalTime {
			return arrivals[i].ArrivalTime < arrivals[j].ArrivalTime
		}
		return arrivals[i].ID < arrivals[j].ID
	})

	for {
		select {
		case t := <-e.taskIn:
			if t == nil {
				continue
			}
			cp := t.Clone()
			cp.Normalize()
			if err := cp.Validate(); err != nil {
				return nil, err
			}
			arrivals = append(arrivals, cp)
			sort.Slice(arrivals, func(i, j int) bool {
				if arrivals[i].ArrivalTime != arrivals[j].ArrivalTime {
					return arrivals[i].ArrivalTime < arrivals[j].ArrivalTime
				}
				return arrivals[i].ID < arrivals[j].ID
			})
		default:
			goto startRun
		}
	}
startRun:

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-e.taskIn:
				if !ok {
					return
				}
				if t == nil {
					continue
				}
				cp := t.Clone()
				cp.Normalize()
				if err := cp.Validate(); err != nil {
					select {
					case errCh <- err:
					default:
					}
					cancel()
					return
				}
				e.mu.Lock()
				arrivals = append(arrivals, cp)
				sort.Slice(arrivals, func(i, j int) bool {
					if arrivals[i].ArrivalTime != arrivals[j].ArrivalTime {
						return arrivals[i].ArrivalTime < arrivals[j].ArrivalTime
					}
					return arrivals[i].ID < arrivals[j].ID
				})
				e.mu.Unlock()
			}
		}
	}()

	simDone := make(chan struct{})
	go func() {
		defer close(simDone)
		e.simulate(ctx, &arrivals)
	}()

	select {
	case err := <-errCh:
		cancel()
		<-simDone
		wg.Wait()
		return nil, err
	case <-simDone:
		cancel()
		close(e.taskIn)
		wg.Wait()
	}

	snap := e.collector.Snapshot()
	return &Result{
		Metrics:     snap,
		Gantt:       append([]GanttSegment(nil), e.gantt...),
		MissedTasks: append([]*task.Task(nil), e.missed...),
		Completed:   append([]*task.Task(nil), e.completed...),
	}, nil
}

type activeJob struct {
	workerID   int
	task       *task.Task
	completeAt int64
}

func (e *Engine) simulate(ctx context.Context, arrivals *[]*task.Task) {
	workerCount := e.cfg.WorkerCount
	active := make([]activeJob, 0, workerCount)
	arrivalIdx := 0
	currentTime := int64(0)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		e.mu.Lock()
		local := *arrivals
		e.mu.Unlock()

		for arrivalIdx < len(local) && local[arrivalIdx].ArrivalTime <= currentTime {
			e.sched.Push(local[arrivalIdx])
			arrivalIdx++
		}
		e.collector.RecordQueueDepth(e.sched.Len())
		e.sched.OnTick(currentTime)

		active = e.finishAtTime(active, currentTime)
		e.dispatch(&active, currentTime, workerCount)

		if arrivalIdx >= len(local) && len(active) == 0 && e.sched.Len() == 0 {
			e.collector.SetSimulationHorizon(currentTime)
			return
		}

		nextArrival := int64(-1)
		if arrivalIdx < len(local) {
			nextArrival = local[arrivalIdx].ArrivalTime
		}
		nextComplete := int64(-1)
		for _, j := range active {
			if nextComplete < 0 || j.completeAt < nextComplete {
				nextComplete = j.completeAt
			}
		}

		nextTime, ok := nextEventTime(currentTime, nextArrival, nextComplete)
		if !ok {
			e.collector.SetSimulationHorizon(currentTime)
			return
		}

		idleWorkers := workerCount - len(active)
		if idleWorkers > 0 && nextTime > currentTime {
			e.collector.AddWorkerIdle((nextTime - currentTime) * int64(idleWorkers))
		}

		currentTime = nextTime
		e.collector.SetSimulationHorizon(currentTime)
	}
}

func nextEventTime(current, arrival, complete int64) (int64, bool) {
	best := int64(-1)
	if arrival >= 0 && arrival > current {
		best = arrival
	}
	if complete >= 0 && complete > current {
		if best < 0 || complete < best {
			best = complete
		}
	}
	return best, best >= 0
}

func (e *Engine) finishAtTime(active []activeJob, t int64) []activeJob {
	var remaining []activeJob
	for _, j := range active {
		if j.completeAt <= t {
			j.task.CompletionTime = j.completeAt
			j.task.Status = task.StatusCompleted
			wait := j.task.StartTime - j.task.ArrivalTime
			if wait < 0 {
				wait = 0
			}
			e.collector.RecordCompleted(wait)
			e.collector.AddWorkerBusy(j.task.Duration)
			e.recordCompleted(j.task, j.workerID)
		} else {
			remaining = append(remaining, j)
		}
	}
	return remaining
}

func (e *Engine) dispatch(active *[]activeJob, currentTime int64, workerCount int) {
	busy := map[int]bool{}
	for _, j := range *active {
		busy[j.workerID] = true
	}

	for len(*active) < workerCount {
		next, ok := e.sched.Pop()
		if !ok {
			return
		}
		e.collector.RecordProcessed()
		workerID := 1
		for id := 1; id <= workerCount; id++ {
			if !busy[id] {
				workerID = id
				break
			}
		}
		busy[workerID] = true

		if !next.FeasibleAt(currentTime) {
			next.Status = task.StatusMissedDeadline
			next.StartTime = currentTime
			next.CompletionTime = currentTime
			next.AssignedWorker = workerID
			e.recordMissed(next, workerID)
			continue
		}

		start := currentTime
		end := start + next.Remaining
		next.Status = task.StatusRunning
		next.StartTime = start
		next.AssignedWorker = workerID
		*active = append(*active, activeJob{workerID: workerID, task: next, completeAt: end})
	}
}

func (e *Engine) recordCompleted(t *task.Task, workerID int) {
	cp := t.Clone()
	e.completed = append(e.completed, cp)
	e.gantt = append(e.gantt, GanttSegment{
		WorkerID:  workerID,
		TaskID:    t.ID,
		TaskName:  t.Name,
		StartMs:   t.StartTime,
		EndMs:     t.CompletionTime,
		Algorithm: string(e.cfg.Algorithm),
		Missed:    false,
	})
}

func (e *Engine) recordMissed(t *task.Task, workerID int) {
	cp := t.Clone()
	e.missed = append(e.missed, cp)
	e.collector.RecordMissed()
	e.gantt = append(e.gantt, GanttSegment{
		WorkerID:  workerID,
		TaskID:    t.ID,
		TaskName:  t.Name,
		StartMs:   t.StartTime,
		EndMs:     t.CompletionTime,
		Algorithm: string(e.cfg.Algorithm),
		Missed:    true,
	})
}

func RunQuick(ctx context.Context, cfg Config, tasks []*task.Task) (*Result, error) {
	eng := New(cfg)
	return eng.Run(ctx, tasks)
}

func FormatMissed(tasks []*task.Task) string {
	if len(tasks) == 0 {
		return ""
	}
	var b []byte
	b = append(b, "\n=== Missed Tasks ===\n"...)
	for _, t := range tasks {
		line := fmt.Sprintf("- %s (Deadline: %dms, Completed: %dms)\n", t.ID, t.Deadline, t.CompletionTime)
		b = append(b, line...)
	}
	return string(b)
}

func SaveResultJSON(path string, res *Result) error {
	type payload struct {
		Metrics     metrics.Snapshot `json:"metrics"`
		Gantt       []GanttSegment   `json:"gantt"`
		MissedTasks []*task.Task     `json:"missed_tasks"`
		Completed   []*task.Task     `json:"completed_tasks"`
	}
	data, err := jsonMarshal(payload{
		Metrics:     res.Metrics,
		Gantt:       res.Gantt,
		MissedTasks: res.MissedTasks,
		Completed:   res.Completed,
	})
	if err != nil {
		return err
	}
	return writeFile(path, data)
}

func jsonMarshal(v any) ([]byte, error) {
	return jsonMarshalImpl(v)
}

func writeFile(path string, data []byte) error {
	return writeFileImpl(path, data)
}
