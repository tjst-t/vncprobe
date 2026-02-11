package vnc

import (
	"fmt"
	"image"
	"image/color"
)

// DiffRatio returns the fraction of pixels that differ between two images.
// Images must have the same dimensions; returns error otherwise.
func DiffRatio(a, b image.Image) (float64, error) {
	ab := a.Bounds()
	bb := b.Bounds()
	if ab.Dx() != bb.Dx() || ab.Dy() != bb.Dy() {
		return 0, fmt.Errorf("image size mismatch: %dx%d vs %dx%d", ab.Dx(), ab.Dy(), bb.Dx(), bb.Dy())
	}

	total := ab.Dx() * ab.Dy()
	if total == 0 {
		return 0, nil
	}

	diff := 0
	for y := ab.Min.Y; y < ab.Max.Y; y++ {
		for x := ab.Min.X; x < ab.Max.X; x++ {
			r1, g1, b1, _ := a.At(x, y).RGBA()
			r2, g2, b2, _ := b.At(x, y).RGBA()
			if !colorsEqual(r1, g1, b1, r2, g2, b2) {
				diff++
			}
		}
	}

	return float64(diff) / float64(total), nil
}

func colorsEqual(r1, g1, b1, r2, g2, b2 uint32) bool {
	return color.RGBA{R: uint8(r1 >> 8), G: uint8(g1 >> 8), B: uint8(b1 >> 8)} ==
		color.RGBA{R: uint8(r2 >> 8), G: uint8(g2 >> 8), B: uint8(b2 >> 8)}
}
