package api

import (
	"net/http"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

// Server agrupa las dependencias que necesitan los handlers HTTP: el
// WorkerPool (para Submit/Lookup), el JobRegistry (para construir el
// Job concreto a partir del JSON que llega en el body) y la
// DeadLetterQueue (para los endpoints de /dlq).
type Server struct {
	wp          *queue.WorkerPool
	jobRegistry *queue.JobRegistry
	dlq         *queue.DeadLetterQueue
}

// NewServer construye el Server con sus dependencias ya armadas
// (típicamente desde cmd/server/main.go).
func NewServer(wp *queue.WorkerPool, jobRegistry *queue.JobRegistry, dlq *queue.DeadLetterQueue) *Server {
	return &Server{wp: wp, jobRegistry: jobRegistry, dlq: dlq}
}

// Routes arma el mux con todos los endpoints. Separado de NewServer
// para poder testear el wiring de rutas de forma explícita, sin tener
// que levantar un listener TCP real (net/http/httptest alcanza).
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", s.handleCreateJob)
	mux.HandleFunc("GET /jobs/{id}", s.handleGetJob)
	mux.HandleFunc("GET /dlq", s.handleListDLQ)
	mux.HandleFunc("POST /dlq/{id}/retry", s.handleRetryDLQ)
	return mux
}