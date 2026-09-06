package queue_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

// flakyJob falla las primeras `failTimes` veces y después tiene éxito.
// Necesita receptor por puntero porque debe recordar estado (calls)
// entre invocaciones sucesivas de Execute.
type flakyJob struct {
	failTimes int
	calls     int
}

func (j *flakyJob) Execute(ctx context.Context) error {
	j.calls++
	if j.calls <= j.failTimes {
		return errors.New("simulated transient failure")
	}
	return nil
}

func TestEndToEnd_QueueLifecycle_SuccessAndDeadLetter(t *testing.T) {
	q := queue.NewQueue()
	dlq := queue.NewDeadLetterQueue()
	ctx := context.Background()

	flaky := &flakyJob{failTimes: 2}
	flakyTask := queue.NewTask(flaky)

	deadEnd := alwaysFailJob{err: errors.New("smtp: connection refused")}
	deadTask := queue.NewTask(deadEnd)

	q.Enqueue(flakyTask)
	q.Enqueue(deadTask)

	// Simula el loop de un worker: procesa hasta vaciar la cola,
	// reencolando cualquier tarea que haya quedado en Retrying.
	for q.Len() > 0 {
		task, err := q.Dequeue()
		if err != nil {
			t.Fatalf("dequeue inesperado: %v", err)
		}

		queue.ProcessTask(ctx, task, dlq)

		switch task.Status {
		case queue.StatusRetrying:
			q.Enqueue(task)
		case queue.StatusCompleted, queue.StatusDead:
			// estado terminal, no vuelve a la cola
		default:
			t.Fatalf("estado inesperado tras ProcessTask: %v", task.Status)
		}
	}

	// La tarea flaky debería haber terminado exitosa, tras 2 fallos.
	if flakyTask.Status != queue.StatusCompleted {
		t.Errorf("esperaba flakyTask Status=%v, obtuve %v", queue.StatusCompleted, flakyTask.Status)
	}
	if flakyTask.Attempts != 2 {
		t.Errorf("esperaba flakyTask Attempts=2, obtuve %d", flakyTask.Attempts)
	}

	// La tarea que siempre falla debería haber terminado en la DLQ.
	if deadTask.Status != queue.StatusDead {
		t.Errorf("esperaba deadTask Status=%v, obtuve %v", queue.StatusDead, deadTask.Status)
	}
	if deadTask.Attempts != queue.MaxAttempts {
		t.Errorf("esperaba deadTask Attempts=%d, obtuve %d", queue.MaxAttempts, deadTask.Attempts)
	}

	got, err := dlq.Remove(deadTask.ID)
	
	if err != nil {
		t.Fatalf("esperaba encontrar deadTask en la DLQ: %v", err)
	}
	if got.ID != deadTask.ID {
		t.Errorf("ID inesperado recuperado de la DLQ: %v", got.ID)
	}
}