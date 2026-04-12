package main

import (
	"log"
	"net"
	"os"

	sensorv1 "github.com/ArthurDescourvieres/BenchLab/grpc-service/gen/benchlab/sensor/v1"
	"github.com/ArthurDescourvieres/BenchLab/internal/sensorsvc"
	"github.com/ArthurDescourvieres/BenchLab/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	addr := ":9090"
	if p := os.Getenv("GRPC_PORT"); p != "" {
		if len(p) > 0 && p[0] == ':' {
			addr = p
		} else {
			addr = ":" + p
		}
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	st := store.NewMemoryStore()
	svc := sensorsvc.New(st)
	sensorv1.RegisterSensorServiceServer(s, newGRPCSensorServer(svc))
	reflection.Register(s)

	log.Printf("gRPC listening on %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
