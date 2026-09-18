package queue

import (
	"context"
	"time"
)

type WorkerPool struct {
	inbox      *Queue
	tasks      chan *Task
	dlq        *DeadLetterQueue
	registry   *TaskRegistry  
	numWorkers int
}

func NewWorkerPool(numWorkers int, channelCapacity int, dlq *DeadLetterQueue) *WorkerPool {
	return &WorkerPool{
		inbox:      NewQueue(),
		tasks:      make(chan *Task, channelCapacity),
		dlq:        dlq,
		registry:   NewTaskRegistry(),   
		numWorkers: numWorkers,
	}
}

func (wp *WorkerPool) Submit(task *Task) {
	wp.registry.Add(task)   
	wp.inbox.Enqueue(task)
}

func (wp *WorkerPool) Pending() int {
	return wp.inbox.Len()
}

func (wp *WorkerPool) Start(ctx context.Context) {
	go wp.dispatch(ctx)
	for i := 0; i < wp.numWorkers; i++ {
		go wp.runWorker(ctx)
	}
}

func (wp *WorkerPool) dispatch(ctx context.Context) {
	for {
		task, err := wp.inbox.DequeueWait(ctx)
		if err != nil {
			return
		}

		select {
		case wp.tasks <- task:
		case <-ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) runWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-wp.tasks:
			ProcessTask(ctx, task, wp.dlq)
			if task.Status == StatusRetrying {
				delay := BackoffDuration(task.Attempts)
				wp.scheduleRetry(ctx, task, delay)
			}
		}
	}
}

func (wp *WorkerPool) scheduleRetry(ctx context.Context, task *Task, delay time.Duration) {
	go func() {
		select {
		case <-time.After(delay):
			wp.Submit(task)
		case <-ctx.Done():
		}
	}()
}

func (wp *WorkerPool) Lookup(id string) (*Task, error) {   
	return wp.registry.Get(id)
}