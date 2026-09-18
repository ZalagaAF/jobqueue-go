package api

import (
	"net/http"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

type retryDLQResponse struct {
	ID string `json:"id"`
}

func (s *Server) handleListDLQ(w http.ResponseWriter, r *http.Request) {
	tasks := s.dlq.List()

	snapshots := make([]queue.TaskSnapshot, 0, len(tasks))
	for _, task := range tasks {
		snapshots = append(snapshots, task.Snapshot())
	}

	writeJSON(w, http.StatusOK, snapshots)
}

func (s *Server) handleRetryDLQ(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	task, err := s.dlq.Remove(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	task.ResetForRetry()
	s.wp.Submit(task)

	writeJSON(w, http.StatusAccepted, retryDLQResponse{ID: task.ID})
}