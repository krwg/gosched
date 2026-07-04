# Plugins

GoSched supports **custom scheduling policies** via Go plugins (`.so` on Linux/macOS).

## Interface

Plugins must export a package variable:

```go
var Scheduler plugin.Factory
```

`plugin.Factory` requires:

```go
type Factory interface {
    Name() string
    Build() scheduler.Scheduler
}
```

## Example plugin

See [`examples/plugin/main.go`](../examples/plugin/main.go).

Build on Linux:

```bash
go build -buildmode=plugin -o deadline-priority.so ./examples/plugin
gosched serve --plugin ./deadline-priority.so
```

## Built-in custom scheduler

Library users can register heap-based policies without `.so` files:

```go
custom := scheduler.NewCustom("my-policy", func(a, b *task.Task) bool {
    return a.Deadline < b.Deadline
})
res, _ := engine.RunQuick(ctx, engine.Config{
    Algorithm: scheduler.EDF,
    WorkerCount: 4,
    Custom: custom,
}, tasks)
```

## Platform notes

| Platform | Plugin support |
|----------|----------------|
| Linux | `.so` via `--plugin` |
| macOS | `.so` via `--plugin` |
| Windows | Use `Custom` scheduler in code (Go plugins do not support `.dll`) |

## Safety

Plugins run in-process. Only load plugins from trusted sources.
