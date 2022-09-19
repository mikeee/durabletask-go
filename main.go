package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/microsoft/durabletask-go/backend"
	"github.com/microsoft/durabletask-go/backend/sqlite"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 4001, "The server port")
var dbFilePath = flag.String("db", "taskhub.sqlite3", "The path to the sqlite file to use (or create if not exists)")
var ctx = context.Background()

func main() {
	// Parse command-line arguments
	flag.Parse()

	grpcServer := grpc.NewServer()
	worker := createTaskHubWorker(grpcServer, *dbFilePath, backend.DefaultLogger())
	if err := worker.Start(ctx); err != nil {
		log.Fatalf("failed to start worker: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Printf("server listening at %v\n", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func createTaskHubWorker(server *grpc.Server, sqliteFilePath string, logger backend.Logger) backend.TaskHubWorker {
	sqliteOptions := sqlite.NewSqliteOptions(sqliteFilePath)
	be := sqlite.NewSqliteBackend(sqliteOptions, logger)
	executor := backend.NewGrpcExecutor(server, be, logger)
	orchestrationWorker := backend.NewOrchestrationWorker(be, executor, logger, backend.NewWorkerOptions())
	activityWorker := backend.NewActivityTaskWorker(be, executor, logger, backend.NewWorkerOptions())
	taskHubWorker := backend.NewTaskHubWorker(be, orchestrationWorker, activityWorker, logger)
	return taskHubWorker
}
