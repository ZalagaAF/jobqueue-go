package queue_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func TestJobRegistry_Build_ConstructsRegisteredJob(t *testing.T) {
	registry := queue.NewJobRegistry()
	registry.Register("email", func(payload json.RawMessage) (queue.Job, error) {
		var j queue.EmailJob
		if err := json.Unmarshal(payload, &j); err != nil {
			return nil, err
		}
		return j, nil
	})

	payload := json.RawMessage(`{"to":"alguien@example.com","subject":"hola"}`)
	job, err := registry.Build("email", payload)
	if err != nil {
		t.Fatalf("no esperaba error, obtuve: %v", err)
	}

	emailJob, ok := job.(queue.EmailJob)
	if !ok {
		t.Fatalf("esperaba un queue.EmailJob, obtuve %T", job)
	}
	if emailJob.To != "alguien@example.com" {
		t.Errorf("esperaba To=alguien@example.com, obtuve %v", emailJob.To)
	}
	if emailJob.Subject != "hola" {
		t.Errorf("esperaba Subject=hola, obtuve %v", emailJob.Subject)
	}
}

func TestJobRegistry_Build_UnknownType_ReturnsError(t *testing.T) {
	registry := queue.NewJobRegistry()

	_, err := registry.Build("no-existe", nil)

	if !errors.Is(err, queue.ErrUnknownJobType) {
		t.Errorf("esperaba queue.ErrUnknownJobType, obtuve: %v", err)
	}
}

func TestJobRegistry_Build_FactoryReturnsError_PropagatesIt(t *testing.T) {
	registry := queue.NewJobRegistry()
	registry.Register("email", func(payload json.RawMessage) (queue.Job, error) {
		var j queue.EmailJob
		if err := json.Unmarshal(payload, &j); err != nil {
			return nil, err
		}
		return j, nil
	})

	_, err := registry.Build("email", json.RawMessage(`{invalid`))

	if err == nil {
		t.Fatal("esperaba un error por JSON inválido, obtuve nil")
	}
}