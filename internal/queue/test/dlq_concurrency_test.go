package queue_test

import (
	"sync"
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func TestDeadLetterQueue_ConcurrentAdd_NoDataRace(t *testing.T) {
	dlq := queue.NewDeadLetterQueue()

	const numGoroutines = 50
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dlq.Add(queue.NewTask(fakeJob{name: "dead"}))
		}()
	}
	wg.Wait()

	if len(dlq.List()) != numGoroutines {
		t.Errorf("esperaba %d tareas en la DLQ, obtuve %d", numGoroutines, len(dlq.List()))
	}
}