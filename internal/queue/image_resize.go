// internal/queue/image_resize.go
package queue

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// ImageResizeJob es un job real (a diferencia de EmailJob, que es
// simulado): abre un archivo de imagen en disco, lo redimensiona con
// nearest neighbor, y guarda el resultado. SourcePath apunta al
// archivo de origen; el archivo de salida se deriva automáticamente
// de ese path (convención, no un campo nuevo): "foto.jpg" produce
// "foto_resized.jpg", sobrescribiendo si ya existía.
type ImageResizeJob struct {
	SourcePath string `json:"source_path"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

func (j ImageResizeJob) Execute(ctx context.Context) error {
	file, err := os.Open(j.SourcePath)
	if err != nil {
		return err
	}
	defer file.Close()

	src, format, err := image.Decode(file)
	if err != nil {
		return err
	}

	resized := resizeNearestNeighbor(src, j.Width, j.Height)

	outputPath := derivedOutputPath(j.SourcePath)
	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	switch format {
	case "jpeg":
		return jpeg.Encode(out, resized, nil)
	case "png":
		return png.Encode(out, resized)
	default:
		return fmt.Errorf("queue: formato de imagen no soportado para guardar: %s", format)
	}
}

// derivedOutputPath calcula el path de salida por convención,
// insertando "_resized" antes de la extensión: "/tmp/foto.jpg" pasa
// a "/tmp/foto_resized.jpg". Si ya existe un archivo con ese nombre,
// Execute lo sobrescribe sin preguntar 
func derivedOutputPath(sourcePath string) string {
	ext := filepath.Ext(sourcePath)
	base := strings.TrimSuffix(sourcePath, ext)
	return base + "_resized" + ext
}

// resizeNearestNeighbor redimensiona src a dstWidth x dstHeight usando
// el algoritmo nearest neighbor: para cada píxel de la imagen de
// salida, copia el color del píxel de la imagen original que le
// queda más cerca según un escalado lineal de coordenadas.
func resizeNearestNeighbor(src image.Image, dstWidth, dstHeight int) image.Image {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))

	for y := 0; y < dstHeight; y++ {
		srcY := srcBounds.Min.Y + y*srcHeight/dstHeight
		for x := 0; x < dstWidth; x++ {
			srcX := srcBounds.Min.X + x*srcWidth/dstWidth
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	return dst
}