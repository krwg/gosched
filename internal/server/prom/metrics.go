package prom

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry = prometheus.NewRegistry()
	factory  = promauto.With(registry)

	SchedulesTotal = factory.NewCounterVec(prometheus.CounterOpts{
		Name: "gosched_schedules_total",
		Help: "Total scheduling runs by algorithm and outcome",
	}, []string{"algorithm", "status"})

	TasksCompleted = factory.NewCounter(prometheus.CounterOpts{
		Name: "gosched_tasks_completed_total",
		Help: "Tasks completed across all runs",
	})

	TasksMissed = factory.NewCounter(prometheus.CounterOpts{
		Name: "gosched_tasks_missed_total",
		Help: "Tasks that missed deadlines across all runs",
	})

	AverageWait = factory.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gosched_average_wait_ms",
		Help: "Average wait time of last scheduling run",
	}, []string{"algorithm"})

	CPUUtilization = factory.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gosched_cpu_utilization_pct",
		Help: "CPU utilization of last scheduling run",
	}, []string{"algorithm"})

	QueueDepth = factory.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gosched_max_queue_depth",
		Help: "Maximum queue depth of last scheduling run",
	}, []string{"algorithm"})

	ActiveRequests = factory.NewGauge(prometheus.GaugeOpts{
		Name: "gosched_active_requests",
		Help: "In-flight schedule requests",
	})
)

var once sync.Once

func Handler() http.Handler {
	once.Do(func() {
		registry.MustRegister(
			prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
			prometheus.NewGoCollector(),
		)
	})
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

func RecordRun(algorithm string, completed, missed int, avgWait, cpuUtil float64, maxDepth int) {
	status := "ok"
	if missed > 0 {
		status = "degraded"
	}
	SchedulesTotal.WithLabelValues(algorithm, status).Inc()
	TasksCompleted.Add(float64(completed))
	TasksMissed.Add(float64(missed))
	AverageWait.WithLabelValues(algorithm).Set(avgWait)
	CPUUtilization.WithLabelValues(algorithm).Set(cpuUtil)
	QueueDepth.WithLabelValues(algorithm).Set(float64(maxDepth))
}
