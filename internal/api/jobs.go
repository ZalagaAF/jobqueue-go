// internal/api/jobs.go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

// createJobRequest es el contrato de entrada de POST /jobs: type le
// dice al JobRegistry qué factory usar, payload es el JSON crudo que
// esa factory va a deserializar al struct concreto (EmailJob, etc.).
// json.RawMessage pospone el parseo del payload hasta que sabemos con
// qué tipo concreto construirlo.
type createJobRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type createJobResponse struct {
	ID string `json:"id"`
}

func (s *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	var req createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo invalido: "+err.Error())
		return
	}

	job, err := s.jobRegistry.Build(req.Type, req.Payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task := queue.NewTask(job)
	s.wp.Submit(task)

	writeJSON(w, http.StatusCreated, createJobResponse{ID: task.ID})
}

// handleGetJob resuelve GET /jobs/{id}. Usa Lookup para encontrar la
// tarea sin importar su estado, y Snapshot() (nunca los campos de
// Task directo) porque un worker podría estar escribiéndola en este
// mismo instante desde otra goroutine.
func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	task, err := s.wp.Lookup(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, task.Snapshot())
}