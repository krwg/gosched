# GoSched API

GoSched exposes scheduling over **HTTP (JSON)** and **gRPC**.

## HTTP

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness probe |
| `GET` | `/ready` | Readiness probe |
| `GET` | `/metrics` | Prometheus metrics |
| `POST` | `/api/v1/schedule` | Run scheduling simulation |

### Schedule request

```json
{
  "algorithm": "edf",
  "workers": 4,
  "tasks": [
    {
      "id": "task-1",
      "name": "Render",
      "duration": 40,
      "deadline": 120,
      "priority": 1,
      "arrival_time": 0
    }
  ]
}
```

### Schedule response

Returns the full engine result: `metrics`, `gantt`, `missed_tasks`, `completed_tasks`.

### Example

```bash
gosched serve --http :8080 --grpc :50051

curl -s http://localhost:8080/api/v1/schedule \
  -H "Content-Type: application/json" \
  -d @tests/fixtures/tasks.json
```

## gRPC

Proto: [`api/proto/gosched/v1/scheduler.proto`](../api/proto/gosched/v1/scheduler.proto)

| RPC | Description |
|-----|-------------|
| `Schedule` | Submit tasks and receive aggregate metrics |
| `Health` | Service health and version |

### Go client

```go
conn, _ := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := rpcv1.NewSchedulerServiceClient(conn)
res, _ := client.Schedule(ctx, &rpcv1.ScheduleRequest{
    Algorithm: "edf",
    Workers:   4,
    Tasks:     []*rpcv1.Task{{Id: "a", Name: "A", Duration: 10, Deadline: 100, Priority: 1}},
})
```

## Prometheus metrics

| Metric | Type | Labels |
|--------|------|--------|
| `gosched_schedules_total` | Counter | `algorithm`, `status` |
| `gosched_tasks_completed_total` | Counter | — |
| `gosched_tasks_missed_total` | Counter | — |
| `gosched_average_wait_ms` | Gauge | `algorithm` |
| `gosched_cpu_utilization_pct` | Gauge | `algorithm` |
| `gosched_max_queue_depth` | Gauge | `algorithm` |
| `gosched_active_requests` | Gauge | — |

Scrape `http://<host>:8080/metrics`.

## Server flags

```bash
gosched serve \
  --http :8080 \
  --grpc :50051 \
  --algorithm edf \
  --workers 4 \
  --plugin ./plugins/custom.so
```
