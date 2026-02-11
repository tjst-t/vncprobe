package vnc

import (
	"image"
	"image/color"
	"sync"
	"testing"
	"time"
)

// sequenceMockClient returns different images on successive Capture() calls.
type sequenceMockClient struct {
	mu     sync.Mutex
	images []image.Image
	index  int
}

func (m *sequenceMockClient) Connect(addr string, password string, timeout time.Duration) error {
	return nil
}

func (m *sequenceMockClient) Capture() (image.Image, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.index >= len(m.images) {
		return m.images[len(m.images)-1], nil
	}
	img := m.images[m.index]
	m.index++
	return img, nil
}

func (m *sequenceMockClient) SendKey(keycode uint32, down bool) error { return nil }
func (m *sequenceMockClient) SendPointer(x, y uint16, buttonMask uint8) error {
	return nil
}
func (m *sequenceMockClient) Close() error { return nil }

func TestWaitForChangeDetectsChange(t *testing.T) {
	red := solidImage(4, 4, color.RGBA{R: 255, A: 255})
	blue := solidImage(4, 4, color.RGBA{B: 255, A: 255})

	client := &sequenceMockClient{images: []image.Image{red, red, blue}}
	opts := WaitOptions{
		Timeout:   2 * time.Second,
		Interval:  10 * time.Millisecond,
		Threshold: 0.01,
	}
	err := WaitForChange(client, opts)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestWaitForChangeTimeout(t *testing.T) {
	red := solidImage(4, 4, color.RGBA{R: 255, A: 255})

	client := &sequenceMockClient{images: []image.Image{red}} // never changes
	opts := WaitOptions{
		Timeout:   100 * time.Millisecond,
		Interval:  10 * time.Millisecond,
		Threshold: 0.01,
	}
	err := WaitForChange(client, opts)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !IsTimeout(err) {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestWaitForStableDetectsStability(t *testing.T) {
	red := solidImage(4, 4, color.RGBA{R: 255, A: 255})
	blue := solidImage(4, 4, color.RGBA{B: 255, A: 255})

	// Changes for first 2 captures, then stays stable
	client := &sequenceMockClient{images: []image.Image{red, blue, blue, blue, blue, blue}}
	opts := WaitOptions{
		Timeout:   2 * time.Second,
		Interval:  10 * time.Millisecond,
		Threshold: 0.01,
	}
	err := WaitForStable(client, opts, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestWaitForStableTimeout(t *testing.T) {
	// Alternating images — never stable
	red := solidImage(4, 4, color.RGBA{R: 255, A: 255})
	blue := solidImage(4, 4, color.RGBA{B: 255, A: 255})

	images := make([]image.Image, 100)
	for i := range images {
		if i%2 == 0 {
			images[i] = red
		} else {
			images[i] = blue
		}
	}

	client := &sequenceMockClient{images: images}
	opts := WaitOptions{
		Timeout:   100 * time.Millisecond,
		Interval:  10 * time.Millisecond,
		Threshold: 0.01,
	}
	err := WaitForStable(client, opts, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !IsTimeout(err) {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}
