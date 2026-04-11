package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
	"time"
)

// créé un pile en go
type MemoryStore struct {
	mu sync.RWMutex
	m  map[string]Sensor
}


func NewMemoryStore() *MemoryStore {
	return &MemoryStore{m: make(map[string]Sensor)}
}

var _ Store = (*MemoryStore)(nil)

func (ms *MemoryStore) Create(ctx context.Context, sensor *Sensor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if sensor == nil {
		return errors.New("sensor is nil")
	}
	id, err := newRandomID()
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	created := *sensor
	created.ID = id
	if created.CreatedAt == "" {
		created.CreatedAt = now
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.m[id] = created
	return nil
}

func (ms *MemoryStore) Get(ctx context.Context, id string) (Sensor, error) {
	if err := ctx.Err(); err != nil {
		return Sensor{}, err
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	sensor, ok := ms.m[id]
	if !ok {
		return Sensor{}, ErrNotFound
	}
	return sensor, nil
}

func (ms *MemoryStore) List(ctx context.Context) ([]Sensor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	out := make([]Sensor, 0, len(ms.m))
	for _, sensor := range ms.m {
		out = append(out, sensor)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (ms *MemoryStore) Update(ctx context.Context, sensor Sensor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	old, ok := ms.m[sensor.ID]
	if !ok {
		return ErrNotFound
	}
	if sensor.CreatedAt == "" {
		sensor.CreatedAt = old.CreatedAt
	}
	ms.m[sensor.ID] = sensor
	return nil
}

func (ms *MemoryStore) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if _, ok := ms.m[id]; !ok {
		return ErrNotFound
	}
	delete(ms.m, id)
	return nil
}

func newRandomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
