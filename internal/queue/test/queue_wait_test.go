package queue_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

// DequeueWait es la versión bloqueante de Dequeue: si la cola está vacía,
// espera (sin polling) hasta que llegue una tarea o el ctx se cancele.
// Es lo que va a usar el dispatcher del WorkerPool para no gastar CPU
// preguntando en loop si hay trabajo nuevo.

func TestDequeueWait_ReturnsImmediately_WhenTaskAlreadyPresent(t *testing.T) {
	q := queue.NewQueue()
	task := queue.NewTask(fakeJob{name: "first"})
	q.Enqueue(task)

	got, err := q.DequeueWait(context.Background())

	if err != nil {
		t.Fatalf("no esperaba error, obtuve: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("esperaba la tarea con ID %v, obtuve %v", task.ID, got.ID)
	}
}

func TestDequeueWait_BlocksUntilTaskArrives(t *testing.T) {
	q := queue.NewQueue()
	ctx := context.Background()

	type result struct {
		task *queue.Task
		err  error
	}
	resultCh := make(chan result, 1)

	go func() {
		task, err := q.DequeueWait(ctx)
		resultCh <- result{task: task, err: err}
	}()

	// Le damos tiempo a la goroutine de arriba para que llegue a
	// bloquearse esperando (si esto fallara, DequeueWait no estaría
	// bloqueando de verdad, sino haciendo polling y devolviendo antes).
	time.Sleep(20 * time.Millisecond)

	task := queue.NewTask(fakeJob{name: "arrived-late"})
	q.Enqueue(task)

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("no esperaba error, obtuve: %v", res.err)
		}
		if res.task.ID != task.ID {
			t.Errorf("esperaba recibir la tarea recién encolada (ID %v), obtuve %v", task.ID, res.task.ID)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("DequeueWait no devolvió la tarea encolada dentro del timeout")
	}
}

func TestDequeueWait_ReturnsWhenContextIsCancelled(t *testing.T) {
	q := queue.NewQueue()
	ctx, cancel := context.WithCancel(context.Background())

	type result struct {
		task *queue.Task
		err  error
	}
	resultCh := make(chan result, 1)

	go func() {
		task, err := q.DequeueWait(ctx)
		resultCh <- result{task: task, err: err}
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case res := <-resultCh:
		if res.task != nil {
			t.Errorf("esperaba task=nil al cancelar, obtuve %v", res.task)
		}
		if !errors.Is(res.err, context.Canceled) {
			t.Errorf("esperaba context.Canceled, obtuve: %v", res.err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("DequeueWait no retornó tras cancelar el contexto")
	}
}