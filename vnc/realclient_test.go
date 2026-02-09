package vnc

import (
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/tjst-t/vncprobe/testutil"
)

func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 60), G: uint8(y * 60), B: 128, A: 255})
		}
	}
	return img
}

func TestRealClientConnectAndClose(t *testing.T) {
	img := testImage()
	srv := testutil.StartFakeVNCServer(t, img)

	client := NewRealClient()
	err := client.Connect(srv.Addr, "", 5*time.Second)
	if err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	defer client.Close()
}

func TestRealClientSendKey(t *testing.T) {
	img := testImage()
	srv := testutil.StartFakeVNCServer(t, img)

	client := NewRealClient()
	if err := client.Connect(srv.Addr, "", 5*time.Second); err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	defer client.Close()

	if err := client.SendKey(0xff0d, true); err != nil {
		t.Fatalf("SendKey error: %v", err)
	}
	if err := client.SendKey(0xff0d, false); err != nil {
		t.Fatalf("SendKey error: %v", err)
	}

	// Verify fake server received the events
	var events []testutil.KeyEvent
	for i := 0; i < 100; i++ {
		events = srv.GetKeyEvents()
		if len(events) >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(events) < 2 {
		t.Fatalf("got %d key events, want >= 2", len(events))
	}
	if events[0].Key != 0xff0d || events[0].DownFlag != true {
		t.Errorf("event[0] = %+v", events[0])
	}
}

func TestRealClientSendPointer(t *testing.T) {
	img := testImage()
	srv := testutil.StartFakeVNCServer(t, img)

	client := NewRealClient()
	if err := client.Connect(srv.Addr, "", 5*time.Second); err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	defer client.Close()

	if err := client.SendPointer(100, 200, 1); err != nil {
		t.Fatalf("SendPointer error: %v", err)
	}

	var events []testutil.PointerEvent
	for i := 0; i < 100; i++ {
		events = srv.GetPointerEvents()
		if len(events) >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(events) < 1 {
		t.Fatalf("got %d pointer events, want >= 1", len(events))
	}
	if events[0].X != 100 || events[0].Y != 200 || events[0].ButtonMask != 1 {
		t.Errorf("event[0] = %+v", events[0])
	}
}

func TestRealClientCapture(t *testing.T) {
	img := testImage()
	srv := testutil.StartFakeVNCServer(t, img)

	client := NewRealClient()
	if err := client.Connect(srv.Addr, "", 5*time.Second); err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	defer client.Close()

	captured, err := client.Capture()
	if err != nil {
		t.Fatalf("Capture error: %v", err)
	}

	bounds := captured.Bounds()
	if bounds.Dx() != 4 || bounds.Dy() != 4 {
		t.Fatalf("captured size = %dx%d, want 4x4", bounds.Dx(), bounds.Dy())
	}
}
