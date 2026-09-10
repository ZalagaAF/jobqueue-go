package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ZalagaAF/jobqueue-go/internal/api"
	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

// alwaysFailJob es un Job que siempre falla. Se registra bajo el tipo
// "always_fail" solo en este archivo de tests — nunca va a producción
// — para poder ejercitar el camino automático de reintento a través
// de la API real, sin depender de un job de negocio concreto.
type alwaysFailJob struct{}

func (alwaysFailJob) Execute(ctx context.Context) error {
	return errors.New("fallo simulado para el test de integracion")
}

// newIntegrationServer arma un Server con workers REALES corriendo
// (wp.Start(ctx)) — a diferencia de newTestServer() en los tests
// unitarios de cada handler, donde nadie llama Start y las tareas se
// quedan quietas en Pending a propósito.
func newIntegrationServer(t *testing.T) (srv *api.Server, wp *queue.WorkerPool, dlq *queue.DeadLetterQueue, cancel func()) {
	t.Helper()

	dlq = queue.NewDeadLetterQueue()
	wp = queue.NewWorkerPool(2, 10, dlq)

	jobRegistry := queue.NewJobRegistry()
	jobRegistry.Register("email", func(payload json.RawMessage) (queue.Job, error) {
		var j queue.EmailJob
		if err := json.Unmarshal(payload, &j); err != nil {
			return nil, err
		}
		return j, nil
	})
	jobRegistry.Register("always_fail", func(payload json.RawMessage) (queue.Job, error) {
		return alwaysFailJob{}, nil
	})

	ctx, cancelCtx := context.WithCancel(context.Background())
	wp.Start(ctx)

	srv = api.NewServer(wp, jobRegistry, dlq)
	return srv, wp, dlq, cancelCtx
}

func postJSON(t *testing.T, srv *api.Server, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	return rec
}

func getJSON(t *testing.T, srv *api.Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	return rec
}

// waitForStatus hace polling sobre GET /jobs/{id} hasta ver el status
// esperado o vencer el timeout. Es un patrón normal en tests de
// integración con concurrencia real: no sabemos en qué milisegundo
// exacto terminará un worker, así que consultamos hasta que pase o
// se acabe el margen.
func waitForStatus(t *testing.T, srv *api.Server, id string, want queue.Status, timeout time.Duration) queue.TaskSnapshot {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var snap queue.TaskSnapshot
	for time.Now().Before(deadline) {
		rec := getJSON(t, srv, "/jobs/"+id)
		if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
			t.Fatalf("no pude parsear la respuesta de GET /jobs/%s: %v", id, err)
		}
		if snap.Status == want {
			return snap
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timeout esperando Status=%v para la tarea %s, quedó en Status=%v", want, id, snap.Status)
	return snap
}

func TestIntegration_SubmittedTask_IsProcessedByRealWorkers_EndsCompleted(t *testing.T) {
	srv, _, _, cancel := newIntegrationServer(t)
	defer cancel()

	rec := postJSON(t, srv, "/jobs", `{"type":"email","payload":{"to":"x@y.com","subject":"hola"}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("esperaba %d, obtuve %d (body: %s)", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var created struct {
		ID string `json:"id"`
	}
	json.Unmarshal(rec.Body.Bytes(), &created)

	waitForStatus(t, srv, created.ID, queue.StatusCompleted, 2*time.Second)
}

func TestIntegration_FailingTask_AutomaticallyMovesToRetrying(t *testing.T) {
	srv, _, _, cancel := newIntegrationServer(t)
	defer cancel()

	rec := postJSON(t, srv, "/jobs", `{"type":"always_fail","payload":{}}`)
	var created struct {
		ID string `json:"id"`
	}
	json.Unmarshal(rec.Body.Bytes(), &created)

	snap := waitForStatus(t, srv, created.ID, queue.StatusRetrying, 2*time.Second)

	if snap.Attempts != 1 {
		t.Errorf("esperaba Attempts=1 tras el primer fallo automático, obtuve %d", snap.Attempts)
	}
}

func TestIntegration_ManualRetryFromDLQ_ThroughRealWorkers_EndsCompleted(t *testing.T) {
	srv, _, dlq, cancel := newIntegrationServer(t)
	defer cancel()

	dead := queue.NewTask(queue.EmailJob{To: "x@y.com"})
	dead.Attempts = queue.MaxAttempts
	dead.Status = queue.StatusDead
	dlq.Add(dead)

	rec := postJSON(t, srv, "/dlq/"+dead.ID+"/retry", "")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("esperaba %d, obtuve %d (body: %s)", http.StatusAccepted, rec.Code, rec.Body.String())
	}

	waitForStatus(t, srv, dead.ID, queue.StatusCompleted, 2*time.Second)

	if len(dlq.List()) != 0 {
		t.Errorf("esperaba que la DLQ quedara vacía tras el retry exitoso, quedan %d", len(dlq.List()))
	}
}