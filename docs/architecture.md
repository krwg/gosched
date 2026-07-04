# Architecture

## Overview

GoSched is a layered real-time task scheduler:

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│ CLI / API   │────▶│ Engine       │────▶│ Worker pool     │
│ gRPC / HTTP │     │ (simulator)  │     │ (goroutines)    │
└─────────────┘     └──────┬───────┘     └─────────────────┘
                           │
                    ┌──────▼───────┐
                    │ Scheduler    │
                    │ RM / EDF /LLF│
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
                    │ Binary heap  │
                    │ O(log n)     │
                    └──────────────┘
```

## Components

| Package | Role |
|---------|------|
| `pkg/task` | Task model, validation, JSON I/O |
| `internal/heap` | Min-heap priority queue |
| `internal/scheduler` | RM, EDF, LLF, custom policies |
| `internal/engine` | Discrete-event simulation, async intake |
| `internal/worker` | Concurrent job execution |
| `internal/metrics` | Run statistics |
| `internal/server/httpserver` | REST + Prometheus |
| `internal/server/grpcserver` | gRPC API |
| `internal/plugin` | `.so` plugin loader |
| `pkg/visualizer` | ASCII + PNG Gantt charts |
| `pkg/rpc/v1` | Protobuf/gRPC contracts |

## Simulation model

Time is simulated in milliseconds. The engine advances a virtual clock to the next **arrival** or **completion** event. Up to `N` workers execute tasks in parallel.

At dispatch:

```
feasible = currentTime + duration <= deadline
```

Infeasible tasks are marked `missed_deadline` immediately.

## Concurrency

- Tasks arrive via seed slice or buffered channel (`Engine.Tasks()`)
- Feeder goroutine merges async submissions
- Simulation goroutine drives scheduling loop
- `gosched serve` runs HTTP and gRPC concurrently with graceful shutdown

## Observability

Prometheus metrics are updated on every API schedule call. CLI runs print human-readable reports and optional JSON artifacts.

## Extension points

1. **Custom heap ordering** — `scheduler.NewCustom`
2. **Go plugins** — `plugin.Load` + `--plugin`
3. **gRPC / HTTP** — integrate with orchestrators, dashboards, CI pipelines

## Further reading

- [Algorithms](algorithms.md)
- [API](api.md)
- [Plugins](plugins.md)
