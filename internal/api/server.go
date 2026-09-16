package api

import (
	"net/http"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)


type Server struct {
	wp          *queue.WorkerPool
	jobRegistry *queue.JobRegistry
	dlq         *queue.DeadLetterQueue
}

func NewServer(wp *queue.WorkerPool, jobRegistry *queue.JobRegistry, dlq *queue.DeadLetterQueue) *Server {
	return &Server{wp: wp, jobRegistry: jobRegistry, dlq: dlq}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", s.handleCreateJob)
	mux.HandleFunc("GET /jobs/{id}", s.handleGetJob)
	mux.HandleFunc("GET /dlq", s.handleListDLQ)
	mux.HandleFunc("POST /dlq/{id}/retry", s.handleRetryDLQ)
	mux.HandleFunc("GET /dashboard", s.handleDashboard)
	return mux
}