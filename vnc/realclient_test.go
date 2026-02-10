package vnc

import (
	"encoding/binary"
	"image"
	"image/color"
	"io"
	"net"
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

// TestRealClientConnectRFB38CompliantServer verifies that the go-vnc client
// can connect to an RFB 3.8 compliant server that sends SecurityResult even
// for SecurityType None. Per §7.1.3, the server MUST send SecurityResult
// regardless of the security type in RFB 3.8.
//
// This test catches the kward/go-vnc bug where securityResultHandshake()
// skips reading SecurityResult for SecurityType None, causing protocol
// desync with compliant servers (QEMU, TigerVNC, libvirt, etc.).
func TestRealClientConnectRFB38CompliantServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	// RFB 3.8 compliant server: sends SecurityResult for SecurityType None.
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Version handshake
		conn.Write([]byte("RFB 003.008\n"))
		ver := make([]byte, 12)
		io.ReadFull(conn, ver)

		// Security: offer SecurityType None (1)
		conn.Write([]byte{1, 1})
		sec := make([]byte, 1)
		io.ReadFull(conn, sec)

		// SecurityResult OK — RFB 3.8 REQUIRES this even for None
		binary.Write(conn, binary.BigEndian, uint32(0))

		// ClientInit
		ci := make([]byte, 1)
		io.ReadFull(conn, ci)

		// ServerInit: 4x4 framebuffer
		binary.Write(conn, binary.BigEndian, uint16(4)) // width
		binary.Write(conn, binary.BigEndian, uint16(4)) // height
		// PixelFormat: 32bpp, depth 24, little-endian, true-color
		conn.Write([]byte{
			32, 24, 0, 1,
			0, 255, 0, 255, 0, 255,
			16, 8, 0,
			0, 0, 0,
		})
		// Desktop name
		binary.Write(conn, binary.BigEndian, uint32(4))
		conn.Write([]byte("test"))

		// Drain client messages to keep connection alive
		buf := make([]byte, 256)
		for {
			if _, err := conn.Read(buf); err != nil {
				return
			}
		}
	}()

	client := NewRealClient()
	err = client.Connect(ln.Addr().String(), "", 5*time.Second)
	if err != nil {
		t.Fatalf("Connect to RFB 3.8 compliant server failed: %v", err)
	}
	defer client.Close()

	// Verify framebuffer dimensions are correct.
	// If the client skips reading SecurityResult (4 bytes of zeros),
	// those bytes are consumed as width=0, height=0 instead, and
	// the rest of ServerInit is parsed from a 4-byte offset — all garbled.
	w := client.conn.FramebufferWidth()
	h := client.conn.FramebufferHeight()
	if w != 4 || h != 4 {
		t.Fatalf("framebuffer size = %dx%d, want 4x4 (protocol desync due to unread SecurityResult?)", w, h)
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
