package queue_test

import (
	"context"
	"image"
	"image/color"
	"image/png"
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

func TestImageResizeJob_Execute_ResizesAndSavesImage(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "foto.png")

	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.RGBA{255, 0, 0, 255})
	src.Set(1, 0, color.RGBA{0, 255, 0, 255})
	src.Set(0, 1, color.RGBA{0, 0, 255, 255})
	src.Set(1, 1, color.RGBA{255, 255, 255, 255})

	f, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("no pude crear el archivo fuente: %v", err)
	}
	if err := png.Encode(f, src); err != nil {
		t.Fatalf("no pude codificar la imagen fuente: %v", err)
	}
	f.Close()

	job := queue.ImageResizeJob{
		SourcePath: sourcePath,
		Width:      8,
		Height:     8,
	}

	if err := job.Execute(context.Background()); err != nil {
		t.Fatalf("Execute devolvio error inesperado: %v", err)
	}

	outputPath := filepath.Join(dir, "foto_resized.png")
	out, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("esperaba que el archivo de salida existiera en %s: %v", outputPath, err)
	}
	defer out.Close()

	decoded, _, err := image.Decode(out)
	if err != nil {
		t.Fatalf("no pude decodificar el archivo de salida: %v", err)
	}

	if decoded.Bounds().Dx() != 8 || decoded.Bounds().Dy() != 8 {
		t.Errorf("dimensiones del resultado = %v, esperaba 8x8", decoded.Bounds())
	}
}