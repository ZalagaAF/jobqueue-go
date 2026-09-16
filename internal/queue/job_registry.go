package queue

import (
	"encoding/json"
	"errors"
	"sync"
)

var ErrUnknownJobType = errors.New("queue: tipo de job desconocido")

// JobFactory construye un Job concreto a partir de su payload JSON.
// Cada tipo de Job (EmailJob, ImageResizeJob, lo que venga después)
// aporta la suya propia.
type JobFactory func(payload json.RawMessage) (Job, error)

// JobRegistry es el lado "Registry" del patrón Strategy+Registry: Job
// ya es Strategy (cualquier struct que implemente Execute sirve), y
// JobRegistry mapea un string (el "type" que llega en el JSON de
// POST /jobs) a la factory que sabe construir ese Job concreto.
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

// Build busca la factory de jobType y la usa para construir el Job a
// partir de payload. Devuelve ErrUnknownJobType si nadie registró ese
// tipo, o el error de la propia factory si el payload es inválido
// (por ejemplo, JSON malformado).
func (r *JobRegistry) Build(jobType string, payload json.RawMessage) (Job, error) {
	r.mu.Lock()
	factory, ok := r.factories[jobType]
	r.mu.Unlock()

	if !ok {
		return nil, ErrUnknownJobType
	}
	return factory(payload)
}