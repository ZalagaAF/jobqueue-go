package queue

import (
	"errors"
	"sync"
)


var ErrTaskNotFound = errors.New("queue: tarea no encontrada")

// DeadLetterQueue almacena tareas que agotaron sus reintentos.
type DeadLetterQueue struct {
	mu    sync.Mutex
	items map[string]*Task
}

func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{items: make(map[string]*Task)}
}

func (d *DeadLetterQueue) Add(task *Task) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.items[task.ID] = task
}

func (d *DeadLetterQueue) List() []*Task {
	d.mu.Lock()
	defer d.mu.Unlock()
	tasks := make([]*Task, 0, len(d.items))
	for _, task := range d.items {
		tasks = append(tasks, task)
	}
	return tasks
}

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