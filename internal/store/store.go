package store

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("sensor not found")

type Sensor struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Location      string  `json:"location"`
	Unit          string  `json:"unit"`
	Status        string  `json:"status"`
	LastValue     float64 `json:"last_value"`
	LastReadingAt string  `json:"last_reading_at"`
	CreatedAt     string  `json:"created_at"`
}

type Store interface {
	Create(ctx context.Context, s *Sensor) error
	Get(ctx context.Context, id string) (Sensor, error)
	List(ctx context.Context) ([]Sensor, error)
	Update(ctx context.Context, s Sensor) error
	Delete(ctx context.Context, id string) error
}