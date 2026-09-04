// internal/queue/queue.go
package queue

import (
	"errors"
	"time"
	"crypto/rand"
	"encoding/hex"
)

const (
	// backoffBase es el delay del primer reintento (attempt=0).
	backoffBase = 1 * time.Second

	// backoffFactor multiplica el delay en cada intento sucesivo.
	backoffFactor = 2

	// MaxBackoffDelay es el techo: ningún delay calculado supera este valor,
	// sin importar cuántos intentos hayan pasado.
	MaxBackoffDelay = 30 * time.Second
)

// ErrEmptyQueue se devuelve cuando se intenta desencolar una tarea
// y no hay ninguna pendiente.
var ErrEmptyQueue = errors.New("queue: no hay tareas pendientes")

// Queue es una cola FIFO en memoria de tareas pendientes de ejecutar.
type Queue struct {
	items []*Task
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
}


// NewQueue crea una cola vacía lista para usar.
func NewQueue() *Queue {
	return &Queue{items: make([]*Task, 0)}
}

// NewTask crea una Task en estado inicial (Pending, sin intentos)
// para el Job dado, con un ID único generado aleatoriamente.
func NewTask(job Job) *Task {
	return &Task{
		ID:     generateID(),
		Job:    job,
		Status: StatusPending,
	}
}

// generateID crea un identificador aleatorio de 16 caracteres hex
// (8 bytes de entropía), suficiente para evitar colisiones en el
// volumen de este proyecto sin depender de librerías externas.
func generateID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		// crypto/rand.Read prácticamente nunca falla en sistemas reales
		// (leería de /dev/urandom o equivalente del SO); si falla, algo
		// está gravemente mal con el sistema operativo, no con la lógica
		// del programa — entrar en pánico es la respuesta correcta acá.
		panic("queue: no se pudo generar ID aleatorio: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// Len devuelve la cantidad de tareas actualmente en la cola.
func (q *Queue) Len() int {
	return len(q.items)
}

// Enqueue agrega una tarea al final de la cola. La Queue no construye
// ni modifica el Task — solo lo almacena — para poder reencolar tanto
// tareas nuevas (via NewTask) como tareas recicladas desde la DLQ.
func (q *Queue) Enqueue(task *Task) {
	q.items = append(q.items, task)
}

// Dequeue retira y devuelve la tarea más antigua de la cola (FIFO).
// Si la cola está vacía, devuelve ErrEmptyQueue.
func (q *Queue) Dequeue() (*Task, error) {
	if len(q.items) == 0 {
		return nil, ErrEmptyQueue
	}

	task := q.items[0]
	q.items = q.items[1:]
	return task, nil
}

// BackoffDuration calcula cuánto esperar antes del siguiente reintento,
// usando backoff exponencial con un techo (MaxBackoffDelay).
// attempt=0 es el delay antes del primer reintento (tras el primer fallo).
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

