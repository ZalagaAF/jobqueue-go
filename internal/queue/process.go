package queue

import "context"

const MaxAttempts = 5

func ProcessTask(ctx context.Context, task *Task, dlq *DeadLetterQueue) {
	task.mu.Lock()
	task.Status = StatusProcessing
	task.mu.Unlock()

	err := task.Job.Execute(ctx)

	if err == nil {
		task.mu.Lock()
		task.Status = StatusCompleted
		task.mu.Unlock()
		return
	}

	task.mu.Lock()
	task.Attempts++
	task.LastError = err
	attempts := task.Attempts

	if attempts >= MaxAttempts {
		task.Status = StatusDead
		task.mu.Unlock()
		dlq.Add(task)
		return
	}

	task.Status = StatusRetrying
	task.mu.Unlock()
}