package queue_test

import (
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func TestWorkerPool_Submit_AddsTaskToInbox(t *testing.T) {
	dlq := queue.NewDeadLetterQueue()
	wp := queue.NewWorkerPool(2, 10, dlq)

	task := queue.NewTask(fakeJob{name: "first"})
	wp.Submit(task)

	if wp.Pending() != 1 {
		t.Errorf("esperaba Pending()=1 tras Submit, obtuve %d", wp.Pending())
	}
}

func TestWorkerPool_Submit_NeverBlocks_EvenWithManyTasks(t *testing.T) {
	dlq := queue.NewDeadLetterQueue()
	wp := queue.NewWorkerPool(2, 10, dlq)

	const totalTasks = 100
	for i := 0; i < totalTasks; i++ {
		wp.Submit(queue.NewTask(fakeJob{name: "burst"}))
	}

	if wp.Pending() != totalTasks {
		t.Errorf("esperaba Pending()=%d, obtuve %d", totalTasks, wp.Pending())
	}
}