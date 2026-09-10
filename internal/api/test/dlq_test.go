package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func TestListDLQ_ReturnsDeadTasks(t *testing.T) {
	srv, _, dlq := newTestServer()

	task := queue.NewTask(queue.EmailJob{To: "x@y.com"})
	dlq.Add(task)

	req := httptest.NewRequest(http.MethodGet, "/dlq", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba status %d, obtuve %d (body: %s)", http.StatusOK, rec.Code, rec.Body.String())
	}

	var snapshots []queue.TaskSnapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snapshots); err != nil {
		t.Fatalf("no pude parsear la respuesta: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("esperaba 1 tarea en la DLQ, obtuve %d", len(snapshots))
	}
	if snapshots[0].ID != task.ID {
		t.Errorf("esperaba ID=%v, obtuve %v", task.ID, snapshots[0].ID)
	}
}

func TestRetryDLQ_ExistingID_ResubmitsAndRemovesFromDLQ(t *testing.T) {
	srv, wp, dlq := newTestServer()

	task := queue.NewTask(queue.EmailJob{To: "x@y.com"})
	task.Attempts = queue.MaxAttempts
	task.Status = queue.StatusDead
	dlq.Add(task)

	req := httptest.NewRequest(http.MethodPost, "/dlq/"+task.ID+"/retry", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("esperaba status %d, obtuve %d (body: %s)", http.StatusAccepted, rec.Code, rec.Body.String())
	}

	if len(dlq.List()) != 0 {
		t.Errorf("esperaba que la tarea saliera de la DLQ, quedan %d", len(dlq.List()))
	}
	if wp.Pending() != 1 {
		t.Errorf("esperaba que la tarea se reencolara en el inbox, Pending()=%d", wp.Pending())
	}

	got, err := wp.Lookup(task.ID)
	if err != nil {
		t.Fatalf("esperaba encontrar la tarea via Lookup tras el retry: %v", err)
	}
	snap := got.Snapshot()
	if snap.Status != queue.StatusPending {
		t.Errorf("esperaba Status=%v tras el reset, obtuve %v", queue.StatusPending, snap.Status)
	}
	if snap.Attempts != 0 {
		t.Errorf("esperaba Attempts=0 tras el reset, obtuve %d", snap.Attempts)
	}
}

func TestRetryDLQ_UnknownID_ReturnsNotFound(t *testing.T) {
	srv, _, _ := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/dlq/id-inexistente/retry", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusNotFound, rec.Code)
	}
}