package queue

import "context"

// Job representa cualquier unidad de trabajo que la cola puede ejecutar.
// El motor de la cola solo conoce este contrato — nunca los detalles
// de EmailJob o ImageResizeJob.
type Job interface {
	Execute(ctx context.Context) error
}

// EmailJob es un job simulado: no manda un email de verdad todavía,
// pero ya tiene la forma completa (campos con tags JSON) para poder
// construirse desde el payload de POST /jobs vía JobRegistry.
type EmailJob struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
}

func (j EmailJob) Execute(ctx context.Context) error {
	// usa j.To, j.Subject directamente, con tipos concretos
	return nil
}