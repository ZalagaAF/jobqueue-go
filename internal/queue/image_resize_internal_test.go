package queue

import (
	"image"
	"image/color"
	"testing"
)

func TestResizeNearestNeighbor_ScalesUpSimpleImage(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.Set(0, 0, color.RGBA{255, 0, 0, 255})     // rojo
	src.Set(1, 0, color.RGBA{0, 255, 0, 255})     // verde
	src.Set(0, 1, color.RGBA{0, 0, 255, 255})     // azul
	src.Set(1, 1, color.RGBA{255, 255, 255, 255}) // blanco

	got := resizeNearestNeighbor(src, 4, 4)

	if got.Bounds().Dx() != 4 || got.Bounds().Dy() != 4 {
		t.Fatalf("dimensiones = %v, esperaba 4x4", got.Bounds())
	}

	wantColor := func(x, y int, want color.Color) {
		t.Helper()
		gr, gg, gb, ga := got.At(x, y).RGBA()
		wr, wg, wb, wa := want.RGBA()
		if gr != wr || gg != wg || gb != wb || ga != wa {
			t.Errorf("pixel (%d,%d) = %v, esperaba %v", x, y, got.At(x, y), want)
		}
	}

	red := color.RGBA{255, 0, 0, 255}
	green := color.RGBA{0, 255, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	white := color.RGBA{255, 255, 255, 255}

	wantColor(0, 0, red)
	wantColor(1, 0, red)
	wantColor(2, 0, green)
	wantColor(3, 0, green)

	wantColor(0, 2, blue)
	wantColor(1, 2, blue)
	wantColor(2, 2, white)
	wantColor(3, 2, white)
}