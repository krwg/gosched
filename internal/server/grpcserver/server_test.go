package grpcserver_test

import (
	"context"
	"net"
	"testing"

	grpcserver "github.com/krwg/gosched/internal/server/grpcserver"
	rpcv1 "github.com/krwg/gosched/pkg/rpc/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCSchedule(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpcserver.New(grpcserver.Options{Workers: 2})
	srv.Register()

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.GracefulStop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := rpcv1.NewSchedulerServiceClient(conn)
	res, err := client.Schedule(context.Background(), &rpcv1.ScheduleRequest{
		Algorithm: "edf",
		Workers:   2,
		Tasks: []*rpcv1.Task{
			{Id: "x", Name: "X", Duration: 10, Deadline: 100, Priority: 1, ArrivalTime: 0},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.GetCompleted() != 1 {
		t.Fatalf("completed=%d", res.GetCompleted())
	}
}

func TestGRPCHealth(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpcserver.New(grpcserver.Options{})
	srv.Register()
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.GracefulStop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := rpcv1.NewSchedulerServiceClient(conn)
	res, err := client.Health(context.Background(), &rpcv1.HealthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if res.GetStatus() != "ok" {
		t.Fatalf("status=%s", res.GetStatus())
	}
}
