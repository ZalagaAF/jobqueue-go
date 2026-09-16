package queue

import "context"

const MaxAttempts = 5

// ProcessTask ejecuta una tarea una vez y actualiza su estado según
// el resultado. No espera el delay de backoff ni reintenta automáticamente
// — eso es responsabilidad de la capa que orquesta el tiempo real (el
// worker). ProcessTask solo decide y refleja el resultado de UN
// intento en el estado de la Task.
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