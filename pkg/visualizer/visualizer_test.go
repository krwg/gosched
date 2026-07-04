package visualizer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krwg/gosched/internal/engine"
	"github.com/krwg/gosched/pkg/visualizer"
)

func TestRenderASCII(t *testing.T) {
	gantt := []engine.GanttSegment{
		{WorkerID: 1, TaskID: "a", TaskName: "A", StartMs: 0, EndMs: 50, Algorithm: "edf"},
		{WorkerID: 2, TaskID: "b", TaskName: "B", StartMs: 10, EndMs: 60, Algorithm: "edf"},
	}
	out := visualizer.RenderASCII(gantt, 100, 20)
	if !strings.Contains(out, "Worker-1") {
		t.Fatal(out)
	}
}

func TestRenderPNG(t *testing.T) {
	gantt := []engine.GanttSegment{
		{WorkerID: 1, TaskID: "a", StartMs: 0, EndMs: 40, Algorithm: "edf"},
	}
	path := filepath.Join(t.TempDir(), "chart.png")
	if err := visualizer.RenderPNG(gantt, 100, path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
