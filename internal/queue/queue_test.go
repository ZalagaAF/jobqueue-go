// internal/queue/queue_test.go
package queue_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

// fakeJob es un stub mínimo: no hace nada real, solo satisface
// la interfaz Job para poder testear la cola sin depender de
// implementaciones concretas como EmailJob.
type fakeJob struct {
	name string
}

func (f fakeJob) Execute(ctx context.Context) error {
	return nil
}

func TestNewQueue_IsEmpty(t *testing.T) {
	q := queue.NewQueue()

	if q.Len() != 0 {
		t.Errorf("esperaba cola vacía (Len=0), obtuve Len=%d", q.Len())
	}
}

func TestEnqueue_IncrementsLen(t *testing.T) {
	q := queue.NewQueue()
	task := queue.NewTask(fakeJob{name: "first"})

	q.Enqueue(task)

	if q.Len() != 1 {
		t.Errorf("esperaba Len=1 tras encolar una tarea, obtuve Len=%d", q.Len())
	}
}

func TestDequeue_OnEmptyQueue_ReturnsError(t *testing.T) {
	q := queue.NewQueue()

	_, err := q.Dequeue()

	if !errors.Is(err, queue.ErrEmptyQueue) {
		t.Errorf("esperaba queue.ErrEmptyQueue, obtuve: %v", err)
	}
}

func TestDequeue_ReturnsTasksInFIFOOrder(t *testing.T) {
	q := queue.NewQueue()
	first := queue.NewTask(fakeJob{name: "first"})
	second := queue.NewTask(fakeJob{name: "second"})

	q.Enqueue(first)
	q.Enqueue(second)

	got, err := q.Dequeue()
	if err != nil {
		t.Fatalf("no esperaba error, obtuve: %v", err)
	}
	if got.ID != first.ID {
		t.Errorf("esperaba desencolar la tarea 'first' primero, obtuve ID: %v", got.ID)
	}

	got, err = q.Dequeue()
	if err != nil {
		t.Fatalf("no esperaba error, obtuve: %v", err)
	}
	if got.ID != second.ID {
		t.Errorf("esperaba desencolar la tarea 'second' segundo, obtuve ID: %v", got.ID)
	}
}

func TestBackoffDuration_DoublesEachAttempt(t *testing.T) {
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 0, want: 1 * time.Second},
		{attempt: 1, want: 2 * time.Second},
		{attempt: 2, want: 4 * time.Second},
		{attempt: 3, want: 8 * time.Second},
	}

	for _, tc := range cases {
		got := queue.BackoffDuration(tc.attempt)
		if got != tc.want {
			t.Errorf("BackoffDuration(%d) = %v, esperaba %v", tc.attempt, got, tc.want)
		}
	}
}

func TestBackoffDuration_CapsAtMaxDelay(t *testing.T) {
	got := queue.BackoffDuration(20)

	if got != queue.MaxBackoffDelay {
		t.Errorf("BackoffDuration(20) = %v, esperaba el tope %v", got, queue.MaxBackoffDelay)
	}
}

func TestNewTask_HasCorrectInitialState(t *testing.T) {
	job := fakeJob{name: "first"}

	task := queue.NewTask(job)

	if task.Job != job {
		t.Errorf("esperaba que Task.Job sea el job original")
	}
	if task.Status != queue.StatusPending {
		t.Errorf("esperaba Status=%v, obtuve %v", queue.StatusPending, task.Status)
	}
	if task.Attempts != 0 {
		t.Errorf("esperaba Attempts=0, obtuve %d", task.Attempts)
	}
	if task.LastError != nil {
		t.Errorf("esperaba LastError=nil, obtuve %v", task.LastError)
	}
	if task.ID == "" {
		t.Errorf("esperaba un ID no vacío")
	}
}

func TestNewTask_GeneratesUniqueIDs(t *testing.T) {
	job := fakeJob{name: "first"}

	taskA := queue.NewTask(job)
	taskB := queue.NewTask(job)

	if taskA.ID == taskB.ID {
		t.Errorf("esperaba IDs distintos, ambos dieron: %v", taskA.ID)
	}
}