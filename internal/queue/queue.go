// internal/queue/queue.go
package queue

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