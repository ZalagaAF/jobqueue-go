// internal/queue/status.go
package queue

// Status representa el estado de una tarea en su ciclo de vida.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"    // falló este intento, puede reintentar
	StatusRetrying   Status = "retrying"
	StatusDead       Status = "dead"      // agotó reintentos → DLQ
)