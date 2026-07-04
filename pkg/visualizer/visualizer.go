package visualizer

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sort"
	"strings"

	"github.com/krwg/gosched/internal/engine"
)

func RenderASCII(gantt []engine.GanttSegment, horizonMs int64, width int) string {
	if width <= 0 {
		width = 40
	}
	if horizonMs <= 0 {
		for _, s := range gantt {
			if s.EndMs > horizonMs {
				horizonMs = s.EndMs
			}
		}
	}
	if horizonMs <= 0 {
		horizonMs = 1
	}

	byWorker := map[int][]engine.GanttSegment{}
	maxWorker := 0
	for _, s := range gantt {
		byWorker[s.WorkerID] = append(byWorker[s.WorkerID], s)
		if s.WorkerID > maxWorker {
			maxWorker = s.WorkerID
		}
	}

	var b strings.Builder
	b.WriteString("=== Gantt Chart ===\n")
	for w := 1; w <= maxWorker; w++ {
		row := make([]rune, width)
		for i := range row {
			row[i] = '░'
		}
		var label string
		for _, s := range byWorker[w] {
			startCol := int(float64(s.StartMs) / float64(horizonMs) * float64(width))
			endCol := int(float64(s.EndMs) / float64(horizonMs) * float64(width))
			if endCol <= startCol {
				endCol = startCol + 1
			}
			if endCol > width {
				endCol = width
			}
			ch := '█'
			if s.Missed {
				ch = '▓'
			}
			for c := startCol; c < endCol && c < width; c++ {
				row[c] = ch
			}
			label = fmt.Sprintf("Task-%s (%s)", s.TaskID, strings.ToUpper(s.Algorithm))
		}
		b.WriteString(fmt.Sprintf("[Worker-%d] |%s| %s\n", w, string(row), label))
	}
	return b.String()
}

func RenderPNG(gantt []engine.GanttSegment, horizonMs int64, path string) error {
	if horizonMs <= 0 {
		for _, s := range gantt {
			if s.EndMs > horizonMs {
				horizonMs = s.EndMs
			}
		}
	}
	if horizonMs <= 0 {
		horizonMs = 1
	}

	byWorker := map[int][]engine.GanttSegment{}
	maxWorker := 0
	for _, s := range gantt {
		byWorker[s.WorkerID] = append(byWorker[s.WorkerID], s)
		if s.WorkerID > maxWorker {
			maxWorker = s.WorkerID
		}
	}
	if maxWorker == 0 {
		maxWorker = 1
	}

	const (
		margin = 40
		rowH   = 36
		imgW   = 800
		chartW = imgW - margin*2
	)

	imgH := margin*2 + maxWorker*rowH
	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	bg := color.RGBA{24, 24, 28, 255}
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			img.Set(x, y, bg)
		}
	}

	colors := []color.RGBA{
		{52, 120, 246, 255},
		{48, 176, 120, 255},
		{246, 166, 52, 255},
		{176, 82, 222, 255},
	}

	for w := 1; w <= maxWorker; w++ {
		y0 := margin + (w-1)*rowH + 8
		for x := margin; x < margin+chartW; x++ {
			for y := y0; y < y0+20; y++ {
				img.Set(x, y, color.RGBA{45, 45, 52, 255})
			}
		}
		for i, s := range byWorker[w] {
			x0 := margin + int(float64(s.StartMs)/float64(horizonMs)*float64(chartW))
			x1 := margin + int(float64(s.EndMs)/float64(horizonMs)*float64(chartW))
			if x1 <= x0 {
				x1 = x0 + 2
			}
			c := colors[i%len(colors)]
			if s.Missed {
				c = color.RGBA{220, 70, 70, 255}
			}
			for x := x0; x < x1 && x < margin+chartW; x++ {
				for y := y0; y < y0+20; y++ {
					img.Set(x, y, c)
				}
			}
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func FromResultFile(input, outputPNG string) (string, error) {
	data, err := os.ReadFile(input)
	if err != nil {
		return "", err
	}
	var payload struct {
		Gantt   []engine.GanttSegment `json:"gantt"`
		Metrics struct {
			TotalSimulatedMs int64 `json:"total_simulated_ms"`
		} `json:"metrics"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	horizon := payload.Metrics.TotalSimulatedMs
	ascii := RenderASCII(payload.Gantt, horizon, 48)
	if outputPNG != "" {
		if err := RenderPNG(payload.Gantt, horizon, outputPNG); err != nil {
			return ascii, err
		}
	}
	return ascii, nil
}

func SortGantt(g []engine.GanttSegment) []engine.GanttSegment {
	out := append([]engine.GanttSegment(nil), g...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].StartMs != out[j].StartMs {
			return out[i].StartMs < out[j].StartMs
		}
		return out[i].WorkerID < out[j].WorkerID
	})
	return out
}
