package queue

import "context"

type Job interface {
	Execute(ctx context.Context) error
}

type EmailJob struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
}

func (j EmailJob) Execute(ctx context.Context) error {
	// usa j.To, j.Subject directamente, con tipos concretos
	return nil //Es simulado, devolvemos nil
}