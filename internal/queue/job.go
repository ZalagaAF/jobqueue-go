package queue

import "context"

// Job representa cualquier unidad de trabajo que la cola puede ejecutar.
// El motor de la cola solo conoce este contrato — nunca los detalles
// de EmailJob o ImageResizeJob.
type Job interface {
	Execute(ctx context.Context) error
}

type EmailJob struct {
	To      string
	Subject string
}

func (j EmailJob) Execute(ctx context.Context) error {
	// usa j.To, j.Subject directamente, con tipos concretos
	return nil
}