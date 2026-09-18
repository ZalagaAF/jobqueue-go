package queue

import (
	"encoding/json"
	"errors"
	"sync"
)

var ErrUnknownJobType = errors.New("queue: tipo de job desconocido")

type JobFactory func(payload json.RawMessage) (Job, error)

type JobRegistry struct {
	mu        sync.Mutex
	factories map[string]JobFactory
}

func NewJobRegistry() *JobRegistry {
	return &JobRegistry{factories: make(map[string]JobFactory)}
}

func (r *JobRegistry) Register(jobType string, factory JobFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[jobType] = factory
}

func (r *JobRegistry) Build(jobType string, payload json.RawMessage) (Job, error) {
	r.mu.Lock()
	factory, ok := r.factories[jobType]
	r.mu.Unlock()

	if !ok {
		return nil, ErrUnknownJobType
	}
	return factory(payload)
}