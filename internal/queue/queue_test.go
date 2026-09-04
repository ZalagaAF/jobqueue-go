// internal/queue/queue_test.go
package queue_test

import (
	"context"
	"errors"
	"testing"

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

	q.Enqueue(fakeJob{name: "first"})

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

func TestDequeue_ReturnsJobsInFIFOOrder(t *testing.T) {
	q := queue.NewQueue()
	first := fakeJob{name: "first"}
	second := fakeJob{name: "second"}

	q.Enqueue(first)
	q.Enqueue(second)

	got, err := q.Dequeue()
	if err != nil {
		t.Fatalf("no esperaba error, obtuve: %v", err)
	}
	if got != first {
		t.Errorf("esperaba desencolar 'first' primero, obtuve: %v", got)
	}

	got, err = q.Dequeue()
	if err != nil {
		t.Fatalf("no esperaba error, obtuve: %v", err)
	}
	if got != second {
		t.Errorf("esperaba desencolar 'second' segundo, obtuve: %v", got)
	}
}