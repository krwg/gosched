package grpcserver

import (
	"context"
	"net"

	"github.com/krwg/gosched/internal/engine"
	"github.com/krwg/gosched/internal/server/prom"
	"github.com/krwg/gosched/internal/scheduler"
	rpcv1 "github.com/krwg/gosched/pkg/rpc/v1"
	"github.com/krwg/gosched/pkg/task"
	"google.golang.org/grpc"
)

type Server struct {
	rpcv1.UnimplementedSchedulerServiceServer
	algo    scheduler.Algorithm
	workers int
	plugin  scheduler.Scheduler
	grpc    *grpc.Server
}

type Options struct {
	Algorithm scheduler.Algorithm
	Workers   int
	Plugin    scheduler.Scheduler
}

func New(opts Options) *Server {
	if opts.Workers <= 0 {
		opts.Workers = 4
	}
	if opts.Algorithm == "" {
		opts.Algorithm = scheduler.EDF
	}
	return &Server{
		algo:    opts.Algorithm,
		workers: opts.Workers,
		plugin:  opts.Plugin,
		grpc:    grpc.NewServer(),
	}
}

func (s *Server) Register() {
	rpcv1.RegisterSchedulerServiceServer(s.grpc, s)
}

func (s *Server) Serve(lis net.Listener) error {
	return s.grpc.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.grpc.GracefulStop()
}

func (s *Server) GRPCServer() *grpc.Server { return s.grpc }

func (s *Server) Schedule(ctx context.Context, req *rpcv1.ScheduleRequest) (*rpcv1.ScheduleResponse, error) {
	prom.ActiveRequests.Inc()
	defer prom.ActiveRequests.Dec()

	algo := s.algo
	if req.GetAlgorithm() != "" {
		parsed, err := scheduler.ParseAlgorithm(req.GetAlgorithm())
		if err != nil {
			return nil, err
		}
		algo = parsed
	}
	workers := s.workers
	if req.GetWorkers() > 0 {
		workers = int(req.GetWorkers())
	}

	tasks := make([]*task.Task, 0, len(req.GetTasks()))
	for _, t := range req.GetTasks() {
		tasks = append(tasks, &task.Task{
			ID:          t.GetId(),
			Name:        t.GetName(),
			Duration:    t.GetDuration(),
			Deadline:    t.GetDeadline(),
			Priority:    int(t.GetPriority()),
			ArrivalTime: t.GetArrivalTime(),
			Period:      t.GetPeriod(),
		})
	}

	cfg := engine.Config{Algorithm: algo, WorkerCount: workers, Custom: s.plugin}
	res, err := engine.RunQuick(ctx, cfg, tasks)
	if err != nil {
		return nil, err
	}

	prom.RecordRun(
		res.Metrics.Algorithm,
		res.Metrics.Completed,
		res.Metrics.MissedDeadlines,
		res.Metrics.AverageWaitMs,
		res.Metrics.CPUUtilization,
		res.Metrics.MaxQueueDepth,
	)

	return &rpcv1.ScheduleResponse{
		Algorithm:        res.Metrics.Algorithm,
		TasksProcessed:   int32(res.Metrics.TasksProcessed),
		Completed:        int32(res.Metrics.Completed),
		MissedDeadlines:  int32(res.Metrics.MissedDeadlines),
		AverageWaitMs:    res.Metrics.AverageWaitMs,
		CpuUtilizationPct: res.Metrics.CPUUtilization,
		MaxQueueDepth:    int32(res.Metrics.MaxQueueDepth),
		TotalSimulatedMs: res.Metrics.TotalSimulatedMs,
	}, nil
}

func (s *Server) Health(context.Context, *rpcv1.HealthRequest) (*rpcv1.HealthResponse, error) {
	return &rpcv1.HealthResponse{Status: "ok", Version: httpserverVersion()}, nil
}

func httpserverVersion() string { return "1.0.0" }
