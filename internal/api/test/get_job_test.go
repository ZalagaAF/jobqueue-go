package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func TestGetJob_ExistingID_ReturnsSnapshot(t *testing.T) {
	srv, _, _ := newTestServer()

	createBody := bytes.NewBufferString(`{"type":"email","payload":{"to":"x@y.com","subject":"hola"}}`)
	createReq := httptest.NewRequest(http.MethodPost, "/jobs", createBody)
	createRec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(createRec, createReq)

	var created struct {
		ID string `json:"id"`
	}
	json.Unmarshal(createRec.Body.Bytes(), &created)

	getReq := httptest.NewRequest(http.MethodGet, "/jobs/"+created.ID, nil)
	getRec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("esperaba status %d, obtuve %d (body: %s)", http.StatusOK, getRec.Code, getRec.Body.String())
	}

	var snap queue.TaskSnapshot
	if err := json.Unmarshal(getRec.Body.Bytes(), &snap); err != nil {
		t.Fatalf("no pude parsear la respuesta: %v", err)
	}
	if snap.ID != created.ID {
		t.Errorf("esperaba ID=%v, obtuve %v", created.ID, snap.ID)
	}
	if snap.Status != queue.StatusPending {
		t.Errorf("esperaba Status=%v, obtuve %v", queue.StatusPending, snap.Status)
	}
}

func TestGetJob_UnknownID_ReturnsNotFound(t *testing.T) {
	srv, _, _ := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/jobs/id-inexistente", nil)
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("esperaba status %d, obtuve %d", http.StatusNotFound, rec.Code)
	}
}