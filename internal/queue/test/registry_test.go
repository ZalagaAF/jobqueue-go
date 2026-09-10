package queue_test

import (
	"errors"
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func TestWorkerPool_Submit_RegistersTaskForLookup(t *testing.T) {
	dlq := queue.NewDeadLetterQueue()
	wp := queue.NewWorkerPool(2, 10, dlq)

	task := queue.NewTask(fakeJob{name: "first"})
	wp.Submit(task)

	got, err := wp.Lookup(task.ID)
	if err != nil {
		t.Fatalf("no esperaba error, obtuve: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("esperaba encontrar la tarea con ID %v, obtuve %v", task.ID, got.ID)
	}
}

func TestWorkerPool_Lookup_UnknownID_ReturnsErrTaskNotFound(t *testing.T) {
	dlq := queue.NewDeadLetterQueue()
	wp := queue.NewWorkerPool(2, 10, dlq)

	_, err := wp.Lookup("id-inexistente")

	if !errors.Is(err, queue.ErrTaskNotFound) {
		t.Errorf("esperaba queue.ErrTaskNotFound, obtuve: %v", err)
	}
}