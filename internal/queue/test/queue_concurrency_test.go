package queue_test

import (
	"sync"
	"testing"
	"time"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

// TestQueue_ConcurrentEnqueueAndDequeue_NoDataRace simula lo que hará el
// worker pool de Semana 2: varias goroutines (workers) llamando Dequeue()
// sobre la MISMA Queue al mismo tiempo, mientras otras goroutines siguen
// encolando tareas nuevas (productores).
//
// La propiedad que nos importa: cada tarea encolada debe ser entregada
// exactamente UNA vez a algún consumer. Si dos consumers reciben la
// misma tarea (o alguna se pierde), la cola no es segura para uso
// concurrente.
//
// Correr con: go test -race ./internal/queue/...
func TestQueue_ConcurrentEnqueueAndDequeue_NoDataRace(t *testing.T) {
	q := queue.NewQueue()

	const numProducers = 5
	const tasksPerProducer = 20
	const totalTasks = numProducers * tasksPerProducer
	const numConsumers = 5

	var producers sync.WaitGroup
	for p := 0; p < numProducers; p++ {
		producers.Add(1)
		go func() {
			defer producers.Done()
			for i := 0; i < tasksPerProducer; i++ {
				q.Enqueue(queue.NewTask(fakeJob{name: "concurrent"}))
			}
		}()
	}

	results := make(chan string, totalTasks)
	stop := make(chan struct{})
	var consumers sync.WaitGroup
	for c := 0; c < numConsumers; c++ {
		consumers.Add(1)
		go func() {
			defer consumers.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				task, err := q.Dequeue()
				if err != nil {
					continue
				}
				results <- task.ID
			}
		}()
	}

	producers.Wait()
	for q.Len() > 0 {
		time.Sleep(time.Millisecond)
	}
	close(stop)
	consumers.Wait()
	close(results)

	seen := make(map[string]bool)
	count := 0
	for id := range results {
		if seen[id] {
			t.Fatalf("la tarea con ID %v fue entregada más de una vez a distintos consumers", id)
		}
		seen[id] = true
		count++
	}
	if count != totalTasks {
		t.Errorf("esperaba %d tareas dequeued en total, obtuve %d", totalTasks, count)
	}
}