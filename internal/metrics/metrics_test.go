package metrics_test

import (
	"testing"

	"github.com/krwg/gosched/internal/metrics"
)

func TestCollectorSnapshot(t *testing.T) {
	c := metrics.New("edf")
	c.RecordProcessed()
	c.RecordCompleted(10)
	c.RecordCompleted(30)
	c.RecordMissed()
	c.RecordQueueDepth(5)
	c.RecordQueueDepth(3)
	c.AddWorkerBusy(100)
	c.AddWorkerIdle(50)
	c.SetSimulationHorizon(200)

	s := c.Snapshot()
	if s.Completed != 2 {
		t.Fatalf("completed=%d", s.Completed)
	}
	if s.MissedDeadlines != 1 {
		t.Fatalf("missed=%d", s.MissedDeadlines)
	}
	if s.AverageWaitMs != 20 {
		t.Fatalf("avg=%f", s.AverageWaitMs)
	}
	if s.MaxQueueDepth != 5 {
		t.Fatalf("depth=%d", s.MaxQueueDepth)
	}
	if s.CPUUtilization < 60 || s.CPUUtilization > 70 {
		t.Fatalf("util=%f", s.CPUUtilization)
	}
}

func TestFormatReport(t *testing.T) {
	out := metrics.FormatReport(metrics.Snapshot{Algorithm: "EDF", Completed: 1})
	if len(out) < 20 {
		t.Fatal("short report")
	}
}
