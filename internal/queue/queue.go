// internal/queue/queue.go
package queue

import (
	"context"
	"errors"
	"sync"
	"time"
	"crypto/rand"
	"encoding/hex"
)

const (
	// backoffBase es el delay del primer reintento (attempt=0).
	backoffBase = 1 * time.Second

	backoffFactor = 2

	// MaxBackoffDelay 
	MaxBackoffDelay = 30 * time.Second
)


var ErrEmptyQueue = errors.New("queue: no hay tareas pendientes")


type Queue struct {
	mu     sync.Mutex
	items  []*Task
	notify chan struct{}
}

// Task envuelve un Job con su metadata de ejecución: estado actual,
// cantidad de intentos realizados, y el motivo del último error
// (relevante una vez que la tarea entra en reintentos o muere en la DLQ).
type Task struct {
	ID        string
	Job       Job
	Status    Status
	Attempts  int
	LastError error
	mu        sync.Mutex
}

// TaskSnapshot es una copia inmutable y segura de leer del estado de
// una Task en un instante dado. LastError se convierte a string
// porque error no serializa a JSON de forma útil, y el snapshot está
// pensado justamente para exponerse por HTTP.
type TaskSnapshot struct {
	ID        string `json:"id"`
	Status    Status `json:"status"`
	Attempts  int    `json:"attempts"`
	LastError string `json:"last_error,omitempty"`
}

func NewQueue() *Queue {
	return &Queue{
		items:  make([]*Task, 0),
		notify: make(chan struct{}, 1),
	}
}

func NewTask(job Job) *Task {
	return &Task{
		ID:     generateID(),
		Job:    job,
		Status: StatusPending,
	}
}

func generateID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic("queue: no se pudo generar ID aleatorio: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

func (q *Queue) Enqueue(task *Task) {
	q.mu.Lock()
	q.items = append(q.items, task)
	q.mu.Unlock()

	// Avisa a quien esté esperando en DequeueWait que hay algo nuevo.
	select {
	case q.notify <- struct{}{}:
	default:
	}
}

func (q *Queue) Dequeue() (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil, ErrEmptyQueue
	}

	task := q.items[0]
	q.items = q.items[1:]
	return task, nil
}

// DequeueWait es la versión bloqueante de Dequeue. Si hay una tarea
// disponible, la devuelve de inmediato. Si la cola está vacía, se
// bloquea (sin polling, sin gastar CPU) hasta que llegue una tarea
// nueva vía Enqueue, o hasta que ctx se cancele.
// El loop es intencional, no un descuido: tras despertar por la señal,
// volvemos a intentar Dequeue() en vez de asumir que hay algo. Esto
// cubre "wakeups espurios" — por ejemplo, si dos goroutines llaman
// DequeueWait a la vez y solo una tarea llegó, la otra se despierta,
// no encuentra nada, y vuelve a esperar. Es el mismo contrato que
// exige sync.Cond.Wait(): nunca confiar en que despertar signifique
// que la condición se cumple, siempre volver a chequearla.
func (q *Queue) DequeueWait(ctx context.Context) (*Task, error) {
	for {
		task, err := q.Dequeue()
		if err == nil {
			return task, nil
		}

		select {
		case <-q.notify:
			// Había (o llegó) una señal: volvemos a intentar Dequeue().
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func BackoffDuration(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	delay := backoffBase
	for i := 0; i < attempt; i++ {
		delay *= backoffFactor
		if delay >= MaxBackoffDelay {
			return MaxBackoffDelay
		}
	}

	return delay
}

func (t *Task) Snapshot() TaskSnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	lastErr := ""
	if t.LastError != nil {
		lastErr = t.LastError.Error()
	}

	return TaskSnapshot{
		ID:        t.ID,
		Status:    t.Status,
		Attempts:  t.Attempts,
		LastError: lastErr,
	}
}

// ResetForRetry reinicia el estado de la tarea para un reintento
// manual desde la DLQ: vuelve a Pending, resetea Attempts a 0 y
// limpia LastError. 
func (t *Task) ResetForRetry() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = StatusPending
	t.Attempts = 0
	t.LastError = nil
}