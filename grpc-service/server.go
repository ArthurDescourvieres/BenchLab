package main

import (
	"context"
	"errors"

	"github.com/ArthurDescourvieres/BenchLab/grpc-service/gen/benchlab/sensor/v1"
	"github.com/ArthurDescourvieres/BenchLab/internal/sensorsvc"
	"github.com/ArthurDescourvieres/BenchLab/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcSensorServer struct {
	sensorv1.UnimplementedSensorServiceServer
	svc *sensorsvc.SensorService
}

func newGRPCSensorServer(svc *sensorsvc.SensorService) *grpcSensorServer {
	return &grpcSensorServer{svc: svc}
}

func (g *grpcSensorServer) CreateSensor(ctx context.Context, req *sensorv1.CreateSensorRequest) (*sensorv1.Sensor, error) {
	if req == nil || req.GetSensor() == nil {
		return nil, status.Error(codes.InvalidArgument, "sensor is required")
	}
	in := payloadFromProto(req.GetSensor(), false)
	created, err := g.svc.Create(ctx, in)
	if err != nil {
		return nil, mapSensorError(err)
	}
	return sensorToProto(created), nil
}

func (g *grpcSensorServer) GetSensor(ctx context.Context, req *sensorv1.SensorId) (*sensorv1.Sensor, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	found, err := g.svc.Get(ctx, req.GetId())
	if err != nil {
		return nil, mapSensorError(err)
	}
	return sensorToProto(found), nil
}

func (g *grpcSensorServer) ListSensors(ctx context.Context, _ *sensorv1.ListSensorsRequest) (*sensorv1.ListSensorsResponse, error) {
	list, err := g.svc.List(ctx)
	if err != nil {
		return nil, mapSensorError(err)
	}
	out := make([]*sensorv1.Sensor, 0, len(list))
	for i := range list {
		out = append(out, sensorToProto(list[i]))
	}
	return &sensorv1.ListSensorsResponse{Sensors: out}, nil
}

func (g *grpcSensorServer) UpdateSensor(ctx context.Context, req *sensorv1.UpdateSensorRequest) (*sensorv1.Sensor, error) {
	if req == nil || req.GetSensor() == nil {
		return nil, status.Error(codes.InvalidArgument, "sensor is required")
	}
	in := payloadFromProto(req.GetSensor(), true)
	updated, err := g.svc.Update(ctx, req.GetId(), in)
	if err != nil {
		return nil, mapSensorError(err)
	}
	return sensorToProto(updated), nil
}

func (g *grpcSensorServer) DeleteSensor(ctx context.Context, req *sensorv1.SensorId) (*sensorv1.DeleteSensorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := g.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, mapSensorError(err)
	}
	return &sensorv1.DeleteSensorResponse{}, nil
}

func mapSensorError(err error) error {
	switch {
	case errors.Is(err, store.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, sensorsvc.ErrInvalidSensorType),
		errors.Is(err, sensorsvc.ErrInvalidStatus),
		errors.Is(err, sensorsvc.ErrInvalidReadingAt),
		errors.Is(err, sensorsvc.ErrInvalidName),
		errors.Is(err, sensorsvc.ErrIDMismatch),
		errors.Is(err, sensorsvc.ErrInvalidID):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func sensorToProto(s store.Sensor) *sensorv1.Sensor {
	return &sensorv1.Sensor{
		Id:            s.ID,
		Name:          s.Name,
		Type:          s.Type,
		Location:      s.Location,
		Unit:          s.Unit,
		Status:        s.Status,
		LastValue:     s.LastValue,
		LastReadingAt: s.LastReadingAt,
		CreatedAt:     s.CreatedAt,
	}
}

func payloadFromProto(p *sensorv1.Sensor, forUpdate bool) sensorsvc.SensorPayload {
	if p == nil {
		return sensorsvc.SensorPayload{}
	}
	id := p.GetId()
	if !forUpdate {
		id = ""
	}
	return sensorsvc.SensorPayload{
		ID:            id,
		Name:          p.GetName(),
		Type:          p.GetType(),
		Location:      p.GetLocation(),
		Unit:          p.GetUnit(),
		Status:        p.GetStatus(),
		LastValue:     p.GetLastValue(),
		LastReadingAt: p.GetLastReadingAt(),
	}
}
