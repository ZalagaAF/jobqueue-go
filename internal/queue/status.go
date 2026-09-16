package queue


type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"    // falló este intento, puede reintentar
	StatusRetrying   Status = "retrying"
	StatusDead       Status = "dead"      // agotó reintentos, va a la DLQ
)