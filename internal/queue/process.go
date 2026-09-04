// internal/queue/process.go
package queue

import "context"

// MaxAttempts es la cantidad máxima de intentos antes de que una
// tarea se considere agotada y se mueva a la Dead-Letter Queue.
const MaxAttempts = 5

// ProcessTask ejecuta una tarea una vez y actualiza su estado según
// el resultado. No espera el delay de backoff ni reintenta automáticamente
// — eso es responsabilidad de la capa que orquesta el tiempo real (el
// worker, en Semana 2). ProcessTask solo decide y refleja el resultado
// de UN intento en el estado de la Task.
func ProcessTask(ctx context.Context, task *Task, dlq *DeadLetterQueue) {
	task.Status = StatusProcessing

	err := task.Job.Execute(ctx)
	if err == nil {
		task.Status = StatusCompleted
		return
	}

	task.Attempts++
	task.LastError = err

	if task.Attempts >= MaxAttempts {
		task.Status = StatusDead
		dlq.Add(task)
		return
	}

	task.Status = StatusRetrying
}