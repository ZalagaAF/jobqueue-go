package queue_test

import (
	"context"
	"os"
	"path/filepath"
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

func TestImageResizeJob_Execute_SourceFileIsNotAValidImage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no-es-imagen.txt")
	if err := os.WriteFile(path, []byte("esto no es una imagen"), 0644); err != nil {
		t.Fatalf("no pude preparar el archivo de prueba: %v", err)
	}

	job := queue.ImageResizeJob{
		SourcePath: path,
		Width:      100,
		Height:     100,
	}

	err := job.Execute(context.Background())

	if err == nil {
		t.Fatal("esperaba un error porque el archivo no es una imagen valida, pero Execute devolvio nil")
	}
}