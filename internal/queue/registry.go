package queue

import "sync"

type TaskRegistry struct {
	mu    sync.Mutex
	items map[string]*Task
}

func NewTaskRegistry() *TaskRegistry {
	return &TaskRegistry{items: make(map[string]*Task)}
}

// Add indexa una tarea por su ID. 
func (r *TaskRegistry) Add(task *Task) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[task.ID] = task
}

// Get busca una tarea por ID. 
func (r *TaskRegistry) Get(id string) (*Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.items[id]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return task, nil
}