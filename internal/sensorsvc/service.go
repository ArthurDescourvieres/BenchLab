package sensorsvc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ArthurDescourvieres/BenchLab/store"
)

var (
	ErrInvalidSensorType = errors.New("type must be TEMPERATURE, PRESSURE or VIBRATION")
	ErrInvalidStatus     = errors.New("status must be ACTIVE, INACTIVE or MAINTENANCE")
	ErrInvalidReadingAt  = errors.New("last_reading_at must be a valid RFC3339 / ISO8601 timestamp")
	ErrInvalidName       = errors.New("name must be non-empty")
	ErrIDMismatch        = errors.New("id in body must match path id or be omitted")
	ErrInvalidID         = errors.New("sensor id is required")
)

var (
	allowedTypes  = map[string]struct{}{"TEMPERATURE": {}, "PRESSURE": {}, "VIBRATION": {}}
	allowedStatus = map[string]struct{}{"ACTIVE": {}, "INACTIVE": {}, "MAINTENANCE": {}}
)

type SensorService struct {
	store store.Store
}

func New(st store.Store) *SensorService {
	return &SensorService{store: st}
}

type SensorPayload struct {
	ID            string
	Name          string
	Type          string
	Location      string
	Unit          string
	Status        string
	LastValue     float64
	LastReadingAt string
}

func (s *SensorService) Create(ctx context.Context, in SensorPayload) (store.Sensor, error) {
	if err := validatePayload(in); err != nil {
		return store.Sensor{}, err
	}
	sensor := &store.Sensor{
		Name:          strings.TrimSpace(in.Name),
		Type:          in.Type,
		Location:      strings.TrimSpace(in.Location),
		Unit:          strings.TrimSpace(in.Unit),
		Status:        in.Status,
		LastValue:     in.LastValue,
		LastReadingAt: in.LastReadingAt,
	}
	if err := s.store.Create(ctx, sensor); err != nil {
		return store.Sensor{}, err
	}
	return s.store.Get(ctx, sensor.ID)
}

func (s *SensorService) Get(ctx context.Context, id string) (store.Sensor, error) {
	if strings.TrimSpace(id) == "" {
		return store.Sensor{}, ErrInvalidID
	}
	return s.store.Get(ctx, id)
}

func (s *SensorService) List(ctx context.Context) ([]store.Sensor, error) {
	list, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return []store.Sensor{}, nil
	}
	return list, nil
}

func (s *SensorService) Update(ctx context.Context, pathID string, in SensorPayload) (store.Sensor, error) {
	if strings.TrimSpace(pathID) == "" {
		return store.Sensor{}, ErrInvalidID
	}
	if in.ID != "" && in.ID != pathID {
		return store.Sensor{}, ErrIDMismatch
	}
	in.ID = pathID
	if err := validatePayload(in); err != nil {
		return store.Sensor{}, err
	}
	up := store.Sensor{
		ID:            pathID,
		Name:          strings.TrimSpace(in.Name),
		Type:          in.Type,
		Location:      strings.TrimSpace(in.Location),
		Unit:          strings.TrimSpace(in.Unit),
		Status:        in.Status,
		LastValue:     in.LastValue,
		LastReadingAt: in.LastReadingAt,
	}
	if err := s.store.Update(ctx, up); err != nil {
		return store.Sensor{}, err
	}
	return s.store.Get(ctx, pathID)
}

func (s *SensorService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidID
	}
	return s.store.Delete(ctx, id)
}

func validatePayload(in SensorPayload) error {
	if strings.TrimSpace(in.Name) == "" {
		return ErrInvalidName
	}
	if _, ok := allowedTypes[in.Type]; !ok {
		return ErrInvalidSensorType
	}
	if _, ok := allowedStatus[in.Status]; !ok {
		return ErrInvalidStatus
	}
	if err := validateRFC3339(in.LastReadingAt); err != nil {
		return ErrInvalidReadingAt
	}
	return nil
}

func validateRFC3339(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("empty")
	}
	if _, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return nil
	}
	_, err := time.Parse(time.RFC3339, s)
	return err
}
