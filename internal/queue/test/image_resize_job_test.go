package queue_test

import (
	"context"
	"testing"

	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func TestImageResizeJob_Execute_SourceFileDoesNotExist(t *testing.T) {
	job := queue.ImageResizeJob{
		SourcePath: "/tmp/no-existe-jobqueue-test-12345.jpg",
		Width:      100,
		Height:     100,
	}

	err := job.Execute(context.Background())

	if err == nil {
		t.Fatal("esperaba un error porque el archivo de origen no existe, pero Execute devolvió nil")
	}
}