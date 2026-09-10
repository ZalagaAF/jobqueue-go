package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/api"
	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func newTestServer() (*api.Server, *queue.WorkerPool, *queue.DeadLetterQueue) {
	dlq := queue.NewDeadLetterQueue()
	wp := queue.NewWorkerPool(2, 10, dlq)
	jobRegistry := queue.NewJobRegistry()
	jobRegistry.Register("email", func(payload json.RawMessage) (queue.Job, error) {
		var j queue.EmailJob
		if err := json.Unmarshal(payload, &j); err != nil {
			return nil, err
		}
		return j, nil
	})
	return api.NewServer(wp, jobRegistry, dlq), wp, dlq
}

func TestPostJobs_ValidEmailJob_ReturnsCreatedWithID(t *testing.T) {
	srv, _, _ := newTestServer()

	body := bytes.NewBufferString(`{"type":"email","payload":{"to":"x@y.com","subject":"hola"}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", body)
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperaba status %d, obtuve %d (body: %s)", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("no pude parsear la respuesta: %v", err)
	}
	if resp.ID == "" {
		t.Error("esperaba un ID no vacío en la respuesta")
	}
}

func TestPostJobs_UnknownType_ReturnsBadRequest(t *testing.T) {
	srv, _, _ := newTestServer()

	body := bytes.NewBufferString(`{"type":"no-existe","payload":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", body)
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusBadRequest, rec.Code)
	}
}

func TestPostJobs_InvalidJSON_ReturnsBadRequest(t *testing.T) {
	srv, _, _ := newTestServer()

	body := bytes.NewBufferString(`{invalid`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", body)
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusBadRequest, rec.Code)
	}
}

func TestPostJobs_SubmittedTask_IsFindableViaLookup(t *testing.T) {
	srv, wp, _ := newTestServer()

	body := bytes.NewBufferString(`{"type":"email","payload":{"to":"x@y.com","subject":"hola"}}`)
	req := httptest.NewRequest(http.MethodPost, "/jobs", body)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	var resp struct {
		ID string `json:"id"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if _, err := wp.Lookup(resp.ID); err != nil {
		t.Errorf("esperaba encontrar la tarea via Lookup, obtuve error: %v", err)
	}
}