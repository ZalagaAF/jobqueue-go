// internal/queue/worker_internal_test.go
package queue

import (
	"context"
	"testing"
	"time"
)

type noopJob struct{}

func (noopJob) Execute(ctx context.Context) error { return nil }

func TestDispatcher_MovesTaskFromInboxToChannel(t *testing.T) {
	wp := NewWorkerPool(2, 10, NewDeadLetterQueue())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wp.Start(ctx)

	task := NewTask(noopJob{})
	wp.Submit(task)

	select {
	case got := <-wp.tasks:
		if got.ID != task.ID {
			t.Errorf("esperaba recibir la tarea con ID %v, obtuve %v", task.ID, got.ID)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("el dispatcher no movió la tarea al channel dentro del timeout")
	}
}

func TestScheduleRetry_ResubmitsTaskAfterDelay(t *testing.T) {
	wp := NewWorkerPool(1, 5, NewDeadLetterQueue())
	task := NewTask(noopJob{})
	task.Status = StatusRetrying
	task.Attempts = 1

	wp.scheduleRetry(context.Background(), task, 20*time.Millisecond)

	if wp.Pending() != 0 {
		t.Errorf("esperaba Pending()=0 antes de que venza el delay, obtuve %d", wp.Pending())
	}

	time.Sleep(50 * time.Millisecond)

	if wp.Pending() != 1 {
		t.Errorf("esperaba Pending()=1 tras el delay (Submit reencoló la tarea), obtuve %d", wp.Pending())
	}
}

func TestScheduleRetry_DoesNotResubmit_WhenContextCancelled(t *testing.T) {
	wp := NewWorkerPool(1, 5, NewDeadLetterQueue())
	ctx, cancel := context.WithCancel(context.Background())
	task := NewTask(noopJob{})

	wp.scheduleRetry(ctx, task, 30*time.Millisecond)
	cancel()

	time.Sleep(60 * time.Millisecond)

	if wp.Pending() != 0 {
		t.Errorf("esperaba Pending()=0: el ctx se canceló antes de que venciera el delay, obtuve %d", wp.Pending())
	}
}

func TestWorkerPool_Worker_ExecutesSubmittedTask(t *testing.T) {
	wp := NewWorkerPool(1, 1, NewDeadLetterQueue())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wp.Start(ctx)

	done := make(chan struct{})
	task := NewTask(&signalJob{done: done})
	wp.Submit(task)

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("el worker no ejecutó la tarea a tiempo")
	}
}

// signalJob avisa por un channel apenas se ejecuta. Se usa en vez de
// leer task.Status después de procesar: leer un campo de Task desde
// el goroutine del test mientras el worker todavía podría estar
// escribiéndolo sería, otra vez, una data race — el mismo problema
// que ya resolvimos en Queue y DeadLetterQueue, pero ahora sobre Task.
// close(done) da una señal segura y libre de esa race.
type signalJob struct {
	done chan struct{}
}

func (j *signalJob) Execute(ctx context.Context) error {
	close(j.done)
	return nil
}

func TestDispatcher_StopsWhenContextIsCancelled(t *testing.T) {
	wp := NewWorkerPool(2, 10, NewDeadLetterQueue())
	ctx, cancel := context.WithCancel(context.Background())
	wp.Start(ctx)

	cancel()
	time.Sleep(50 * time.Millisecond)

	task := NewTask(noopJob{})
	wp.Submit(task)

	select {
	case <-wp.tasks:
		t.Fatal("el dispatcher siguió moviendo tareas tras cancelar el contexto")
	case <-time.After(100 * time.Millisecond):
	}

	if wp.Pending() != 1 {
		t.Errorf("esperaba Pending()=1 (tarea varada en el inbox), obtuve %d", wp.Pending())
	}
}