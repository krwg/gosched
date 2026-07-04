# Deployment

## Docker Compose (recommended)

Runs **GoSched**, **Prometheus**, and **Grafana** together.

```bash
docker compose up --build -d
```

| Service | URL | Credentials |
|---------|-----|-------------|
| GoSched HTTP | http://localhost:8080 | — |
| GoSched gRPC | localhost:50051 | — |
| Prometheus | http://localhost:9090 | — |
| Grafana | http://localhost:3000 | `admin` / `gosched` |

### Smoke test

```bash
curl -s http://localhost:8080/health
curl -s -X POST http://localhost:8080/api/v1/schedule \
  -H "Content-Type: application/json" \
  -d @tests/fixtures/tasks.json
```

Open Grafana → **GoSched** folder → **GoSched** dashboard. Metrics appear after at least one schedule request.

### Stop

```bash
docker compose down
docker compose down -v
```

## Docker image only

```bash
docker build -t gosched:local .
docker run --rm -p 8080:8080 -p 50051:50051 gosched:local
```

## Production notes

- Put a reverse proxy (nginx, Caddy) in front of HTTP/gRPC
- Change Grafana admin password via `GF_SECURITY_ADMIN_PASSWORD`
- Tune Prometheus retention and scrape interval for your workload
- Plugins (`.so`) require mounting into the container on Linux hosts
