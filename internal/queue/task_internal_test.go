package queue

import (
	"context"
	"sync"
	"testing"
	"time"
)

type slowJob struct{}

func (slowJob) Execute(ctx context.Context) error {
	time.Sleep(10 * time.Millisecond)
	return nil
}

func TestTask_Snapshot_NoDataRaceWithProcessTask(t *testing.T) {
	task := NewTask(slowJob{})
	dlq := NewDeadLetterQueue()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		ProcessTask(context.Background(), task, dlq)
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			_ = task.Snapshot()
			time.Sleep(time.Millisecond)
		}
	}()

	wg.Wait()

	got := task.Snapshot()
	if got.Status != StatusCompleted {
		t.Errorf("esperaba Status=%v, obtuve %v", StatusCompleted, got.Status)
	}
	if got.ID != task.ID {
		t.Errorf("esperaba ID=%v, obtuve %v", task.ID, got.ID)
	}
}