package vnc

import (
	"image"
	"image/color"
	"testing"
)

func solidImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestDiffRatioIdenticalImages(t *testing.T) {
	img := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	ratio, err := DiffRatio(img, img)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ratio != 0.0 {
		t.Errorf("DiffRatio = %f, want 0.0", ratio)
	}
}

func TestDiffRatioCompletelyDifferent(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	b := solidImage(10, 10, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	ratio, err := DiffRatio(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ratio != 1.0 {
		t.Errorf("DiffRatio = %f, want 1.0", ratio)
	}
}

func TestDiffRatioPartialDifference(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	b := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Copy a, then change top-left quarter (25 out of 100 pixels)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			b.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			b.Set(x, y, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		}
	}
	ratio, err := DiffRatio(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ratio != 0.25 {
		t.Errorf("DiffRatio = %f, want 0.25", ratio)
	}
}

func TestDiffRatioSizeMismatch(t *testing.T) {
	a := solidImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	b := solidImage(20, 20, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	_, err := DiffRatio(a, b)
	if err == nil {
		t.Error("expected error for size mismatch, got nil")
	}
}
