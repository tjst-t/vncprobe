package session

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"
	"time"

	"github.com/tjst-t/vncprobe/vnc"
)

type mockVNCClient struct {
	captureImg image.Image
}

func (m *mockVNCClient) Connect(addr string, password string, timeout time.Duration) error {
	return nil
}
func (m *mockVNCClient) Capture() (image.Image, error) {
	return m.captureImg, nil
}
func (m *mockVNCClient) SendKey(keycode uint32, down bool) error     { return nil }
func (m *mockVNCClient) SendPointer(x, y uint16, mask uint8) error   { return nil }
func (m *mockVNCClient) Close() error                                { return nil }

var _ vnc.VNCClient = &mockVNCClient{}

func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}
	return img
}

func TestServerStartAndStop(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "test.sock")
	client := &mockVNCClient{captureImg: testImage()}

	srv := NewServer(client, sock, 0)
	go srv.ListenAndServe()
	defer srv.Shutdown()

	// Wait for server to start
	time.Sleep(50 * time.Millisecond)

	// Send stop command
	c := NewClient(sock)
	err := c.Execute("session", []string{"stop"})
	if err != nil {
		t.Fatalf("execute stop: %v", err)
	}
}

func TestServerExecuteCapture(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "test.sock")
	client := &mockVNCClient{captureImg: testImage()}

	srv := NewServer(client, sock, 0)
	go srv.ListenAndServe()
	defer srv.Shutdown()

	time.Sleep(50 * time.Millisecond)

	c := NewClient(sock)
	outPath := filepath.Join(t.TempDir(), "out.png")
	err := c.Execute("capture", []string{"-o", outPath})
	if err != nil {
		t.Fatalf("execute capture: %v", err)
	}
}

func TestServerExecuteKeyAndType(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "test.sock")
	client := &mockVNCClient{captureImg: testImage()}

	srv := NewServer(client, sock, 0)
	go srv.ListenAndServe()
	defer srv.Shutdown()

	time.Sleep(50 * time.Millisecond)

	c := NewClient(sock)

	if err := c.Execute("key", []string{"enter"}); err != nil {
		t.Fatalf("execute key: %v", err)
	}

	if err := c.Execute("type", []string{"hello"}); err != nil {
		t.Fatalf("execute type: %v", err)
	}
}

func TestServerIdleTimeout(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "test.sock")
	client := &mockVNCClient{captureImg: testImage()}

	srv := NewServer(client, sock, 200*time.Millisecond)

	done := make(chan error, 1)
	go func() {
		done <- srv.ListenAndServe()
	}()

	// Server should auto-shutdown after idle timeout
	select {
	case <-done:
		// OK - server stopped
	case <-time.After(2 * time.Second):
		srv.Shutdown()
		t.Fatal("server did not auto-shutdown within expected time")
	}
}

func TestClientConnectionRefused(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "nonexistent.sock")
	c := NewClient(sock)
	err := c.Execute("key", []string{"enter"})
	if err == nil {
		t.Fatal("expected error connecting to nonexistent socket")
	}
}
