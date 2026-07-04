package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcserver "github.com/krwg/gosched/internal/server/grpcserver"
	httpserver "github.com/krwg/gosched/internal/server/httpserver"
	"github.com/krwg/gosched/internal/scheduler"
	"github.com/krwg/gosched/internal/plugin"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	var (
		httpAddr string
		grpcAddr string
		algorithm string
		workers   int
		pluginPath string
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP (REST + Prometheus) and gRPC servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			algo, err := scheduler.ParseAlgorithm(algorithm)
			if err != nil {
				return err
			}

			var custom scheduler.Scheduler
			if pluginPath != "" {
				f, err := plugin.Load(pluginPath)
				if err != nil {
					return err
				}
				custom = f.Build()
				fmt.Fprintf(os.Stderr, "loaded plugin: %s\n", f.Name())
			}

			httpSrv := httpserver.New(httpserver.Options{
				Addr:      httpAddr,
				Algorithm: algo,
				Workers:   workers,
				Plugin:    custom,
			})
			grpcSrv := grpcserver.New(grpcserver.Options{
				Algorithm: algo,
				Workers:   workers,
				Plugin:    custom,
			})
			grpcSrv.Register()

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			httpErr := make(chan error, 1)
			go func() {
				fmt.Fprintf(os.Stderr, "HTTP listening on %s (metrics: /metrics)\n", httpAddr)
				httpErr <- httpSrv.ListenAndServe()
			}()

			grpcLis, err := net.Listen("tcp", grpcAddr)
			if err != nil {
				return err
			}
			go func() {
				fmt.Fprintf(os.Stderr, "gRPC listening on %s\n", grpcAddr)
				_ = grpcSrv.Serve(grpcLis)
			}()

			select {
			case <-ctx.Done():
			case err := <-httpErr:
				if err != nil && err != http.ErrServerClosed {
					return err
				}
			}

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			grpcSrv.GracefulStop()
			return httpSrv.Shutdown(shutdownCtx)
		},
	}

	cmd.Flags().StringVar(&httpAddr, "http", ":8080", "HTTP listen address")
	cmd.Flags().StringVar(&grpcAddr, "grpc", ":50051", "gRPC listen address")
	cmd.Flags().StringVar(&algorithm, "algorithm", "edf", "Default scheduling algorithm")
	cmd.Flags().IntVar(&workers, "workers", 4, "Default worker pool size")
	cmd.Flags().StringVar(&pluginPath, "plugin", "", "Path to .so scheduler plugin (Linux/macOS)")
	return cmd
}
