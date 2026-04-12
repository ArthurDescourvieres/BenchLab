package main

import (
	"context"
	"fmt"
	"os"
	"time"

	sensorv1 "github.com/ArthurDescourvieres/BenchLab/grpc-service/gen/benchlab/sensor/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := "localhost:9090"
	if len(os.Args) > 1 && os.Args[1] != "" {
		addr = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	cli := sensorv1.NewSensorServiceClient(conn)
	resp, err := cli.CreateSensor(ctx, &sensorv1.CreateSensorRequest{
		Sensor: &sensorv1.Sensor{
			Name:          "Bench-Setup",
			Type:          "TEMPERATURE",
			Location:      "Lab",
			Unit:          "°C",
			Status:        "ACTIVE",
			LastValue:     21.5,
			LastReadingAt: "2026-01-15T10:00:00Z",
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(resp.GetId())
}
