package queue_test

import (
	"errors"
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func TestDeadLetterQueue_AddAndList(t *testing.T) {
	dlq := queue.NewDeadLetterQueue()
	task := queue.NewTask(fakeJob{name: "first"})

	dlq.Add(task)

	got := dlq.List()
	if len(got) != 1 {
		t.Fatalf("esperaba 1 tarea en la DLQ, obtuve %d", len(got))
	}
	if got[0].ID != task.ID {
		t.Errorf("esperaba la tarea con ID %v, obtuve %v", task.ID, got[0].ID)
	}
}

func TestDeadLetterQueue_Remove_ExistingID(t *testing.T) {
	dlq := queue.NewDeadLetterQueue()
	task := queue.NewTask(fakeJob{name: "first"})
	dlq.Add(task)

	removed, err := dlq.Remove(task.ID)

	if err != nil {
		t.Fatalf("no esperaba error, obtuve: %v", err)
	}
	if removed.ID != task.ID {
		t.Errorf("esperaba remover la tarea con ID %v, obtuve %v", task.ID, removed.ID)
	}
	if len(dlq.List()) != 0 {
		t.Errorf("esperaba DLQ vacía tras remover, quedaron %d", len(dlq.List()))
	}
}

func TestDeadLetterQueue_Remove_UnknownID(t *testing.T) {
	dlq := queue.NewDeadLetterQueue()

	_, err := dlq.Remove("id-inexistente")

	if !errors.Is(err, queue.ErrTaskNotFound) {
		t.Errorf("esperaba queue.ErrTaskNotFound, obtuve: %v", err)
	}
}