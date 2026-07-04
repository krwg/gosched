package plugin

import (
	"fmt"
	"plugin"

	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
)

type Factory interface {
	Name() string
	Build() scheduler.Scheduler
}

type LessFunc = func(a, b *task.Task) bool

type HeapFactory struct {
	NameValue string
	Less      LessFunc
}

func (h HeapFactory) Name() string { return h.NameValue }

func (h HeapFactory) Build() scheduler.Scheduler {
	return scheduler.NewCustom(h.NameValue, h.Less)
}

func Load(path string) (Factory, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	sym, err := p.Lookup("Scheduler")
	if err != nil {
		return nil, err
	}
	f, ok := sym.(Factory)
	if !ok {
		return nil, fmt.Errorf("plugin symbol Scheduler must implement plugin.Factory")
	}
	return f, nil
}
