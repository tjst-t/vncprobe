package testdata_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestGenerateExpectedPNG(t *testing.T) {
	const path = "expected.png"
	if _, err := os.Stat(path); err == nil {
		t.Skip("expected.png already exists")
	}

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 128, A: 255})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
