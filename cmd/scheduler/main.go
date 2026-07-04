package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/krwg/gosched/internal/engine"
	"github.com/krwg/gosched/internal/metrics"
	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/pkg/task"
	"github.com/krwg/gosched/pkg/visualizer"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gosched",
	Short: "GoSched — real-time task scheduler (RM, EDF, LLF)",
	Long:  "GoSched is a production-oriented library and CLI for priority-based real-time task scheduling with custom heap, worker pool, and metrics.",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(scheduleCmd())
	rootCmd.AddCommand(visualizeCmd())
	rootCmd.AddCommand(benchmarkCmd())
	rootCmd.AddCommand(serveCmd())
}

func scheduleCmd() *cobra.Command {
	var (
		algorithm  string
		tasksFile  string
		workers    int
		outputJSON string
	)

	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Run scheduling simulation on a task set",
		RunE: func(cmd *cobra.Command, args []string) error {
			algo, err := scheduler.ParseAlgorithm(algorithm)
			if err != nil {
				return err
			}
			tasks, err := task.LoadTasksFromFile(tasksFile)
			if err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			res, err := engine.RunQuick(ctx, engine.Config{
				Algorithm:   algo,
				WorkerCount: workers,
			}, tasks)
			if err != nil {
				return err
			}

			fmt.Print(metrics.FormatReport(res.Metrics))
			fmt.Print(visualizer.RenderASCII(res.Gantt, res.Metrics.TotalSimulatedMs, 48))
			fmt.Print(engine.FormatMissed(res.MissedTasks))

			if outputJSON != "" {
				if err := engine.SaveResultJSON(outputJSON, res); err != nil {
					return err
				}
				fmt.Printf("\nWrote %s\n", outputJSON)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&algorithm, "algorithm", "edf", "Scheduling algorithm: rm, edf, llf")
	cmd.Flags().StringVar(&tasksFile, "tasks", "tasks.json", "Path to tasks JSON file")
	cmd.Flags().IntVar(&workers, "workers", 4, "Worker pool size")
	cmd.Flags().StringVar(&outputJSON, "output", "", "Optional path to write result JSON")
	return cmd
}

func visualizeCmd() *cobra.Command {
	var (
		input  string
		output string
	)

	cmd := &cobra.Command{
		Use:   "visualize",
		Short: "Render Gantt chart from a result JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ascii, err := visualizer.FromResultFile(input, output)
			if err != nil {
				return err
			}
			fmt.Print(ascii)
			if output != "" {
				fmt.Printf("PNG saved to %s\n", output)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&input, "input", "result.json", "Result JSON from schedule --output")
	cmd.Flags().StringVar(&output, "output", "chart.png", "Output PNG path")
	return cmd
}

func benchmarkCmd() *cobra.Command {
	var (
		algorithms string
		iterations int
		taskCount  int
	)

	cmd := &cobra.Command{
		Use:   "benchmark",
		Short: "Compare scheduling algorithms",
		RunE: func(cmd *cobra.Command, args []string) error {
			names := strings.Split(algorithms, ",")
			ctx := context.Background()

			type row struct {
				Algorithm      string
				TotalMs        int64
				Missed         int
				AvgWait        float64
				SchedulingNs   int64
			}
			var rows []row

			for _, name := range names {
				algo, err := scheduler.ParseAlgorithm(strings.TrimSpace(name))
				if err != nil {
					return err
				}
				var totalMs int64
				var missed int
				var waitSum float64
				var schedNs int64

				for i := 0; i < iterations; i++ {
					tasks := syntheticTasks(taskCount, i)
					start := time.Now()
					res, err := engine.RunQuick(ctx, engine.Config{Algorithm: algo, WorkerCount: 4}, tasks)
					schedNs += time.Since(start).Nanoseconds()
					if err != nil {
						return err
					}
					totalMs += res.Metrics.TotalSimulatedMs
					missed += res.Metrics.MissedDeadlines
					waitSum += res.Metrics.AverageWaitMs
				}

				rows = append(rows, row{
					Algorithm:    string(algo),
					TotalMs:      totalMs / int64(iterations),
					Missed:       missed / iterations,
					AvgWait:      waitSum / float64(iterations),
					SchedulingNs: schedNs / int64(iterations),
				})
			}

			fmt.Println("=== Benchmark Results ===")
			for _, r := range rows {
				fmt.Printf("%s: avg_sim=%dms missed=%d avg_wait=%.1fms scheduling=%.2fµs/iter\n",
					strings.ToUpper(r.Algorithm), r.TotalMs, r.Missed, r.AvgWait, float64(r.SchedulingNs)/1000.0)
			}

			data, _ := json.MarshalIndent(rows, "", "  ")
			fmt.Printf("\n%s\n", data)
			return nil
		},
	}

	cmd.Flags().StringVar(&algorithms, "algorithms", "rm,edf,llf", "Comma-separated algorithms")
	cmd.Flags().IntVar(&iterations, "iterations", 1000, "Iterations per algorithm")
	cmd.Flags().IntVar(&taskCount, "tasks", 50, "Synthetic tasks per iteration")
	return cmd
}

func syntheticTasks(n, seed int) []*task.Task {
	out := make([]*task.Task, 0, n)
	for i := 0; i < n; i++ {
		arrival := int64((i*7 + seed*3) % 200)
		duration := int64(10 + (i*5+seed)%40)
		deadline := arrival + duration + int64(30+(i*11+seed)%120)
		out = append(out, &task.Task{
			ID:          fmt.Sprintf("bench-%d-%d", seed, i),
			Name:        fmt.Sprintf("Job-%d", i),
			Duration:    duration,
			Deadline:    deadline,
			Priority:    1 + (i % 10),
			ArrivalTime: arrival,
		})
	}
	return out
}
