package api

import (
	"encoding/json"
	"net/http"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

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

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	task, err := s.wp.Lookup(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, task.Snapshot())
}