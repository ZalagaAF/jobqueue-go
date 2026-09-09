// internal/queue/dlq.go
package queue

import (
	"errors"
	"sync"
)

// ErrTaskNotFound se devuelve cuando se busca una tarea por ID
// y no existe en la colección consultada.
var ErrTaskNotFound = errors.New("queue: tarea no encontrada")

// DeadLetterQueue almacena tareas que agotaron sus reintentos.
// A diferencia de Queue (FIFO estricta), permite listar todo su
// contenido y remover una tarea específica por ID — porque el caso
// de uso es "consultar y reintentar manualmente", no "procesar en orden".
type DeadLetterQueue struct {
	mu    sync.Mutex
	items map[string]*Task
}

// NewDeadLetterQueue crea una DLQ vacía lista para usar.
func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{items: make(map[string]*Task)}
}

// Add mueve una tarea a la DLQ.
func (d *DeadLetterQueue) Add(task *Task) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.items[task.ID] = task
}

// List devuelve todas las tareas actualmente en la DLQ.
// El orden no está garantizado (ver nota de diseño sobre map).
func (d *DeadLetterQueue) List() []*Task {
	d.mu.Lock()
	defer d.mu.Unlock()
	tasks := make([]*Task, 0, len(d.items))
	for _, task := range d.items {
		tasks = append(tasks, task)
	}
	return tasks
}

// Remove retira y devuelve la tarea con el ID dado.
// Si no existe, devuelve ErrTaskNotFound.
func (d *DeadLetterQueue) Remove(id string) (*Task, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	task, ok := d.items[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	delete(d.items, id)
	return task, nil
}