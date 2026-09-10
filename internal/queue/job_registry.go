package queue

import (
	"encoding/json"
	"errors"
	"sync"
)

// ErrUnknownJobType se devuelve cuando se pide construir un Job de un
// tipo que nadie registró.
var ErrUnknownJobType = errors.New("queue: tipo de job desconocido")

// JobFactory construye un Job concreto a partir de su payload JSON.
// Cada tipo de Job (EmailJob, ImageResizeJob, lo que venga después)
// aporta la suya propia.
type JobFactory func(payload json.RawMessage) (Job, error)

// JobRegistry es el lado "Registry" del patrón Strategy+Registry: Job
// ya es Strategy (cualquier struct que implemente Execute sirve), y
// JobRegistry mapea un string (el "type" que llega en el JSON de
// POST /jobs) a la factory que sabe construir ese Job concreto.
//
// Deliberadamente no conoce EmailJob ni ningún tipo concreto — eso se
// registra afuera, en la composición final (cmd/server/main.go).
// Agregar un tipo de Job nuevo el día de mañana significa escribir el
// struct + una llamada a Register; nada en este archivo, ni en los
// handlers HTTP, cambia.
type JobRegistry struct {
	mu        sync.Mutex
	factories map[string]JobFactory
}

// NewJobRegistry crea un registro vacío listo para usar.
func NewJobRegistry() *JobRegistry {
	return &JobRegistry{factories: make(map[string]JobFactory)}
}

// Register asocia un tipo (string) con la factory que sabe construirlo.
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