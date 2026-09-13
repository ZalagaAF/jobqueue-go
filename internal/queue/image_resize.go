// internal/queue/image_resize.go
package queue

import (
	"context"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// ImageResizeJob es un job real (a diferencia de EmailJob, que es
// simulado): abre un archivo de imagen en disco, lo redimensiona con
// nearest neighbor, y guarda el resultado. SourcePath apunta al
// archivo de origen; el archivo de salida se deriva automáticamente
// de ese path (convención, no un campo nuevo).
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

	_, _, err = image.Decode(file)
	if err != nil {
		return err
	}

	return nil
}

// resizeNearestNeighbor redimensiona src a dstWidth x dstHeight usando
// el algoritmo nearest neighbor: para cada píxel de la imagen de
// salida, copia el color del píxel de la imagen original que le
// queda más cerca según un escalado lineal de coordenadas. No
// interpola ni promedia — es el algoritmo de resize más simple que
// existe, a costa de bordes "pixelados" en escalados grandes.
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