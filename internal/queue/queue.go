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
// Es segura para uso concurrente: el worker pool de Semana 2 lanza
// varias goroutines que llaman Enqueue/Dequeue sobre la misma instancia,
// y mu serializa el acceso a items para que dos goroutines nunca lean
// o escriban el slice al mismo tiempo.
type Queue struct {
	mu     sync.Mutex
	items  []*Task
	notify chan struct{}
}

// Task envuelve un Job con su metadata de ejecución: estado actual,
// cantidad de intentos realizados, y el motivo del último error
// (relevante una vez que la tarea entra en reintentos o muere en la DLQ).
//
// mu protege Status, Attempts y LastError: ProcessTask los escribe
// desde la goroutine de un worker, y cualquier lector externo (por
// ejemplo, el handler HTTP de GET /jobs/:id) puede leerlos al mismo
// tiempo desde otra goroutine. Usar Snapshot() en vez de leer los
// campos directamente es lo que evita esa data race.
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

// NewQueue crea una cola vacía lista para usar.
func NewQueue() *Queue {
	return &Queue{
		items:  make([]*Task, 0),
		notify: make(chan struct{}, 1),
	}
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
// También toma el lock: leer len(q.items) mientras otra goroutine
// hace append o reslice es en sí mismo un acceso concurrente sin
// proteger, aunque Len() no escriba nada.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Enqueue agrega una tarea al final de la cola. La Queue no construye
// ni modifica el Task — solo lo almacena — para poder reencolar tanto
// tareas nuevas (via NewTask) como tareas recicladas desde la DLQ.
func (q *Queue) Enqueue(task *Task) {
	q.mu.Lock()
	q.items = append(q.items, task)
	q.mu.Unlock()

	// Avisa a quien esté esperando en DequeueWait que hay algo nuevo.
	// select+default es clave: si ya hay una señal pendiente sin
	// consumir (buffer lleno), no bloquea ni se apila una segunda —
	// una sola señal alcanza para que el que espera vuelva a mirar
	// la cola completa, no importa cuántas tareas se agregaron.
	select {
	case q.notify <- struct{}{}:
	default:
	}
}

// Dequeue retira y devuelve la tarea más antigua de la cola (FIFO).
// Si la cola está vacía, devuelve ErrEmptyQueue.
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
//
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

// Snapshot devuelve el estado actual de la tarea de forma segura para
// llamar desde cualquier goroutine, incluso mientras un worker la está
// procesando en simultáneo.
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
// limpia LastError. Sin esto, reencolar una tarea que ya estaba en
// MaxAttempts la mandaría de vuelta a la DLQ apenas fallara una vez
// más — un reintento manual debería darle un ciclo fresco completo
// de intentos, no uno solo.
func (t *Task) ResetForRetry() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = StatusPending
	t.Attempts = 0
	t.LastError = nil
}