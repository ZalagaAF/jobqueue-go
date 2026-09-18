package queue_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

type alwaysFailJob struct {
	err error
}

func (j alwaysFailJob) Execute(ctx context.Context) error {
	return j.err
}

func TestProcessTask_OnSuccess_MarksCompleted(t *testing.T) {
	task := queue.NewTask(fakeJob{name: "first"})
	dlq := queue.NewDeadLetterQueue()

	queue.ProcessTask(context.Background(), task, dlq)

	if task.Status != queue.StatusCompleted {
		t.Errorf("esperaba Status=%v, obtuve %v", queue.StatusCompleted, task.Status)
	}
	if len(dlq.List()) != 0 {
		t.Errorf("no esperaba tareas en la DLQ, hay %d", len(dlq.List()))
	}
}

func TestProcessTask_OnFailure_BelowMaxAttempts_MarksRetrying(t *testing.T) {
	wantErr := errors.New("smtp: connection refused")
	task := queue.NewTask(alwaysFailJob{err: wantErr})
	dlq := queue.NewDeadLetterQueue()

	queue.ProcessTask(context.Background(), task, dlq)

	if task.Status != queue.StatusRetrying {
		t.Errorf("esperaba Status=%v, obtuve %v", queue.StatusRetrying, task.Status)
	}
	if task.Attempts != 1 {
		t.Errorf("esperaba Attempts=1, obtuve %d", task.Attempts)
	}
	if !errors.Is(task.LastError, wantErr) {
		t.Errorf("esperaba LastError=%v, obtuve %v", wantErr, task.LastError)
	}
	if len(dlq.List()) != 0 {
		t.Errorf("no esperaba tareas en la DLQ todavía, hay %d", len(dlq.List()))
	}
}

func TestProcessTask_OnFailure_ReachesMaxAttempts_MovesToDLQ(t *testing.T) {
	wantErr := errors.New("smtp: connection refused")
	task := queue.NewTask(alwaysFailJob{err: wantErr})
	dlq := queue.NewDeadLetterQueue()

	for i := 0; i < queue.MaxAttempts; i++ {
		queue.ProcessTask(context.Background(), task, dlq)
	}

	if task.Status != queue.StatusDead {
		t.Errorf("esperaba Status=%v, obtuve %v", queue.StatusDead, task.Status)
	}
	if task.Attempts != queue.MaxAttempts {
		t.Errorf("esperaba Attempts=%d, obtuve %d", queue.MaxAttempts, task.Attempts)
	}

	got, err := dlq.Remove(task.ID)
	if err != nil {
		t.Fatalf("esperaba encontrar la tarea en la DLQ, obtuve error: %v", err)
	}
	if !errors.Is(got.LastError, wantErr) {
		t.Errorf("esperaba que la tarea en la DLQ conserve LastError=%v, obtuve %v", wantErr, got.LastError)
	}
}