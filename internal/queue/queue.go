// internal/queue/queue.go
package queue

import "errors"

// ErrEmptyQueue se devuelve cuando se intenta desencolar una tarea
// y no hay ninguna pendiente.
var ErrEmptyQueue = errors.New("queue: no hay tareas pendientes")

// Queue es una cola FIFO en memoria de tareas pendientes de ejecutar.
type Queue struct {
	items []Job
}

// NewQueue crea una cola vacía lista para usar.
func NewQueue() *Queue {
	return &Queue{items: make([]Job, 0)}
}

// Len devuelve la cantidad de tareas actualmente en la cola.
func (q *Queue) Len() int {
	return len(q.items)
}

// Enqueue agrega una tarea al final de la cola.
func (q *Queue) Enqueue(job Job) {
	q.items = append(q.items, job)
}

// Dequeue retira y devuelve la tarea más antigua de la cola (FIFO).
// Si la cola está vacía, devuelve ErrEmptyQueue.
func (q *Queue) Dequeue() (Job, error) {
	if len(q.items) == 0 {
		return nil, ErrEmptyQueue
	}

	job := q.items[0]
	q.items = q.items[1:]
	return job, nil
}

