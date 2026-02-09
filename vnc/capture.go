package vnc

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
)

func SaveImagePNG(w io.Writer, img image.Image) error {
	return png.Encode(w, img)
}

func CaptureToFile(client VNCClient, path string) error {
	img, err := client.Capture()
	if err != nil {
		return fmt.Errorf("capture: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	if err := SaveImagePNG(f, img); err != nil {
		return fmt.Errorf("save PNG %s: %w", path, err)
	}
	return nil
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}
