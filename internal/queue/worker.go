package queue

import (
	"context"
	"time"
)
// WorkerPool coordina el procesamiento concurrente de tareas: recibe
// trabajo vía Submit (sin límite, nunca bloquea al llamador) y lo
// distribuye entre un número fijo de workers a través de un channel
// acotado — el desacople que diagramamos: inbox ilimitado -> dispatcher
// -> channel bounded -> workers.
type WorkerPool struct {
	inbox      *Queue
	tasks      chan *Task
	dlq        *DeadLetterQueue
	registry   *TaskRegistry  
	numWorkers int
}

// NewWorkerPool crea un pool listo para recibir trabajo con Submit,
// pero todavía sin ninguna goroutine corriendo — separar "construir"
// de "arrancar" permite testear el struct sin efectos secundarios
// concurrentes, y deja a quien lo usa decidir cuándo empieza el trabajo real
func NewWorkerPool(numWorkers int, channelCapacity int, dlq *DeadLetterQueue) *WorkerPool {
	return &WorkerPool{
		inbox:      NewQueue(),
		tasks:      make(chan *Task, channelCapacity),
		dlq:        dlq,
		registry:   NewTaskRegistry(),   
		numWorkers: numWorkers,
	}
}

// Submit encola una tarea para ser procesada. Nunca bloquea al
// llamador, sin importar cuán ocupados estén los workers: la tarea
// simplemente se suma al inbox ilimitado, y el dispatcher será quien
// la mueva al channel bounded cuando haya lugar.
func (wp *WorkerPool) Submit(task *Task) {
	wp.registry.Add(task)   
	wp.inbox.Enqueue(task)
}

// Pending devuelve cuántas tareas están esperando en el inbox,
// todavía sin haber sido tomadas por el dispatcher. 
func (wp *WorkerPool) Pending() int {
	return wp.inbox.Len()
}

// Start pone en marcha el pool: lanza el dispatcher como una goroutine
// separada y retorna de inmediato 
// ctx controla el apagado: cuando se cancela, el dispatcher deja de
// mover tareas del inbox al channel y su goroutine termina.
func (wp *WorkerPool) Start(ctx context.Context) {
	go wp.dispatch(ctx)
	for i := 0; i < wp.numWorkers; i++ {
		go wp.runWorker(ctx)
	}
}

// dispatch es el loop que mueve tareas del inbox (ilimitado) al
// channel (bounded). Usa DequeueWait para no hacer polling mientras
// el inbox está vacío, y un select con ctx.Done() en el envío al
// channel para no quedar bloqueado para siempre si el pool se apaga
// justo cuando el channel está lleno.
func (wp *WorkerPool) dispatch(ctx context.Context) {
	for {
		task, err := wp.inbox.DequeueWait(ctx)
		if err != nil {
			return
		}

		select {
		case wp.tasks <- task:
		case <-ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) runWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-wp.tasks:
			ProcessTask(ctx, task, wp.dlq)
			if task.Status == StatusRetrying {
				delay := BackoffDuration(task.Attempts)
				wp.scheduleRetry(ctx, task, delay)
			}
		}
	}
}

// scheduleRetry espera delay (calculado afuera, vía BackoffDuration,
// para que este método sea testeable con delays cortos sin depender
// del backoff real) y luego reencola la tarea con Submit — el mismo
// camino que recorre cualquier tarea nueva, sin lógica especial para
// "tareas que vuelven".
//
// Corre en su propia goroutine: si esperáramos el delay dentro de
// runWorker directamente, ese worker quedaría inutilizado (sin poder
// tomar ninguna otra tarea del channel) durante todo el backoff,
// hasta 30s con MaxBackoffDelay — desperdiciando uno de los N workers
// disponibles mientras no hace nada más que dormir.
func (wp *WorkerPool) scheduleRetry(ctx context.Context, task *Task, delay time.Duration) {
	go func() {
		select {
		case <-time.After(delay):
			wp.Submit(task)
		case <-ctx.Done():
		}
	}()
}

func (wp *WorkerPool) Lookup(id string) (*Task, error) {   // NUEVO
	return wp.registry.Get(id)
}