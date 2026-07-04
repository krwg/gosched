package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Snapshot struct {
	Algorithm        string  `json:"algorithm"`
	TasksProcessed   int     `json:"tasks_processed"`
	Completed        int     `json:"completed"`
	MissedDeadlines  int     `json:"missed_deadlines"`
	AverageWaitMs    float64 `json:"average_wait_ms"`
	CPUUtilization   float64 `json:"cpu_utilization_pct"`
	MaxQueueDepth    int     `json:"max_queue_depth"`
	TotalSimulatedMs int64   `json:"total_simulated_ms"`
	WorkerIdleMs     int64   `json:"worker_idle_ms"`
	WorkerBusyMs     int64   `json:"worker_busy_ms"`
}

type Collector struct {
	mu sync.Mutex

	algorithm    string
	processed    int
	completed    int
	missed       int
	waitSum      int64
	maxQueue     int
	totalSimMs   int64
	workerIdleMs int64
	workerBusyMs int64
}

func New(algorithm string) *Collector {
	return &Collector{algorithm: algorithm}
}

func (c *Collector) SetAlgorithm(name string) {
	c.mu.Lock()
	c.algorithm = name
	c.mu.Unlock()
}

func (c *Collector) RecordQueueDepth(depth int) {
	c.mu.Lock()
	if depth > c.maxQueue {
		c.maxQueue = depth
	}
	c.mu.Unlock()
}

func (c *Collector) RecordProcessed() {
	c.mu.Lock()
	c.processed++
	c.mu.Unlock()
}

func (c *Collector) RecordCompleted(waitMs int64) {
	c.mu.Lock()
	c.completed++
	c.waitSum += waitMs
	c.mu.Unlock()
}

func (c *Collector) RecordMissed() {
	c.mu.Lock()
	c.missed++
	c.mu.Unlock()
}

func (c *Collector) AddSimulationTime(ms int64) {
	c.mu.Lock()
	c.totalSimMs += ms
	c.mu.Unlock()
}

func (c *Collector) SetSimulationHorizon(ms int64) {
	c.mu.Lock()
	if ms > c.totalSimMs {
		c.totalSimMs = ms
	}
	c.mu.Unlock()
}

func (c *Collector) AddWorkerBusy(ms int64) {
	c.mu.Lock()
	c.workerBusyMs += ms
	c.mu.Unlock()
}

func (c *Collector) AddWorkerIdle(ms int64) {
	c.mu.Lock()
	c.workerIdleMs += ms
	c.mu.Unlock()
}

func (c *Collector) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	avgWait := 0.0
	if c.completed > 0 {
		avgWait = float64(c.waitSum) / float64(c.completed)
	}

	util := 0.0
	denom := c.workerBusyMs + c.workerIdleMs
	if denom > 0 {
		util = float64(c.workerBusyMs) / float64(denom) * 100
	}

	return Snapshot{
		Algorithm:        c.algorithm,
		TasksProcessed:   c.processed,
		Completed:        c.completed,
		MissedDeadlines:  c.missed,
		AverageWaitMs:    avgWait,
		CPUUtilization:   util,
		MaxQueueDepth:    c.maxQueue,
		TotalSimulatedMs: c.totalSimMs,
		WorkerIdleMs:     c.workerIdleMs,
		WorkerBusyMs:     c.workerBusyMs,
	}
}

func WriteJSON(path string, snap Snapshot) error {
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func FormatReport(snap Snapshot) string {
	return fmt.Sprintf(`=== Task Scheduler Results ===
Algorithm: %s
Tasks processed: %d
Completed: %d
Missed deadlines: %d
CPU Utilization: %.1f%%
Average wait time: %.0fms
Max queue depth: %d
`,
		snap.Algorithm,
		snap.TasksProcessed,
		snap.Completed,
		snap.MissedDeadlines,
		snap.CPUUtilization,
		snap.AverageWaitMs,
		snap.MaxQueueDepth,
	)
}
