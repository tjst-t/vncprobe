package vnc

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestSaveImagePNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	var buf bytes.Buffer
	err := SaveImagePNG(&buf, img)
	if err != nil {
		t.Fatalf("SaveImagePNG error: %v", err)
	}

	decoded, err := png.Decode(&buf)
	if err != nil {
		t.Fatalf("png.Decode error: %v", err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 4 || bounds.Dy() != 4 {
		t.Fatalf("decoded size = %dx%d, want 4x4", bounds.Dx(), bounds.Dy())
	}
	r, g, b, a := decoded.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
		t.Errorf("pixel (0,0) = (%d,%d,%d,%d), want (255,0,0,255)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestCaptureToFile(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 100, G: 200, B: 50, A: 255})

	mock := &mockClient{captureImage: img}

	outPath := t.TempDir() + "/out.png"
	err := CaptureToFile(mock, outPath)
	if err != nil {
		t.Fatalf("CaptureToFile error: %v", err)
	}

	decoded, err := loadPNG(outPath)
	if err != nil {
		t.Fatalf("loadPNG error: %v", err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 2 {
		t.Fatalf("decoded size = %dx%d, want 2x2", bounds.Dx(), bounds.Dy())
	}
}
