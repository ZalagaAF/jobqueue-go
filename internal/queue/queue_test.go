package queue_test

import (
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func TestNewQueue_IsEmpty(t *testing.T) {
	q := queue.NewQueue()

	if q.Len() != 0 {
		t.Errorf("esperaba cola vacía (Len=0), obtuve Len=%d", q.Len())
	}
}