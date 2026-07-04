package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/krwg/gosched/internal/engine"
	"github.com/krwg/gosched/internal/server/prom"
	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
)

const Version = "1.0.0"

type Server struct {
	addr    string
	plugin  scheduler.Scheduler
	workers int
	algo    scheduler.Algorithm
	srv     *http.Server
}

type Options struct {
	Addr      string
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
		addr:    opts.Addr,
		plugin:  opts.Plugin,
		workers: opts.Workers,
		algo:    opts.Algorithm,
	}
}

type scheduleRequest struct {
	Algorithm string        `json:"algorithm"`
	Workers   int           `json:"workers"`
	Tasks     []*task.Task    `json:"tasks"`
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /ready", s.ready)
	mux.Handle("GET /metrics", prom.Handler())
	mux.HandleFunc("POST /api/v1/schedule", s.schedule)
	return mux
}

func (s *Server) ListenAndServe() error {
	s.srv = &http.Server{
		Addr:              s.addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok", "version": Version})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ready"})
}

func (s *Server) schedule(w http.ResponseWriter, r *http.Request) {
	prom.ActiveRequests.Inc()
	defer prom.ActiveRequests.Dec()

	var req scheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	algo := s.algo
	if req.Algorithm != "" {
		parsed, err := scheduler.ParseAlgorithm(req.Algorithm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		algo = parsed
	}
	workers := s.workers
	if req.Workers > 0 {
		workers = req.Workers
	}

	cfg := engine.Config{Algorithm: algo, WorkerCount: workers, Custom: s.plugin}
	res, err := engine.RunQuick(r.Context(), cfg, req.Tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	prom.RecordRun(
		res.Metrics.Algorithm,
		res.Metrics.Completed,
		res.Metrics.MissedDeadlines,
		res.Metrics.AverageWaitMs,
		res.Metrics.CPUUtilization,
		res.Metrics.MaxQueueDepth,
	)
	writeJSON(w, res)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var running atomic.Bool

func SetRunning(v bool) { running.Store(v) }

func IsRunning() bool { return running.Load() }
