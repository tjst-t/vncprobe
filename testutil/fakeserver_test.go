package testutil

import (
	"encoding/binary"
	"image"
	"image/color"
	"io"
	"net"
	"testing"
	"time"
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

func TestFakeServerHandshake(t *testing.T) {
	srv := StartFakeVNCServer(t, testImage())

	conn, err := net.Dial("tcp", srv.Addr)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	// Read server protocol version
	buf := make([]byte, 12)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read version error: %v", err)
	}
	if string(buf) != "RFB 003.008\n" {
		t.Fatalf("version = %q, want %q", string(buf), "RFB 003.008\n")
	}

	// Send client protocol version
	if _, err := conn.Write([]byte("RFB 003.008\n")); err != nil {
		t.Fatalf("write version error: %v", err)
	}

	// Read security types (1 type: None=1)
	secBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, secBuf); err != nil {
		t.Fatalf("read security types error: %v", err)
	}
	if secBuf[0] != 1 || secBuf[1] != 1 {
		t.Fatalf("security types = %v, want [1, 1]", secBuf)
	}

	// Send selected security type (None=1)
	if _, err := conn.Write([]byte{1}); err != nil {
		t.Fatalf("write security type error: %v", err)
	}

	// No SecurityResult for security type None (kward/go-vnc compatibility).

	// Send ClientInit (shared=1)
	if _, err := conn.Write([]byte{1}); err != nil {
		t.Fatalf("write client init error: %v", err)
	}

	// Read ServerInit
	var width, height uint16
	if err := binary.Read(conn, binary.BigEndian, &width); err != nil {
		t.Fatalf("read width error: %v", err)
	}
	if err := binary.Read(conn, binary.BigEndian, &height); err != nil {
		t.Fatalf("read height error: %v", err)
	}
	if width != 4 || height != 4 {
		t.Fatalf("size = %dx%d, want 4x4", width, height)
	}
}

func TestFakeServerRecordsKeyEvent(t *testing.T) {
	srv := StartFakeVNCServer(t, testImage())
	conn := doHandshake(t, srv.Addr)
	defer conn.Close()

	// Send KeyEvent: msg-type=4, down-flag=1, padding=0,0, key=0xff0d
	keyMsg := []byte{4, 1, 0, 0, 0, 0, 0xff, 0x0d}
	if _, err := conn.Write(keyMsg); err != nil {
		t.Fatalf("write key event error: %v", err)
	}

	// Send key release
	keyMsg2 := []byte{4, 0, 0, 0, 0, 0, 0xff, 0x0d}
	if _, err := conn.Write(keyMsg2); err != nil {
		t.Fatalf("write key event 2 error: %v", err)
	}

	// Wait for server to process
	var events []KeyEvent
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
		t.Errorf("event[0] = %+v, want {Key:0xff0d, DownFlag:true}", events[0])
	}
	if events[1].Key != 0xff0d || events[1].DownFlag != false {
		t.Errorf("event[1] = %+v, want {Key:0xff0d, DownFlag:false}", events[1])
	}
}

func TestFakeServerRecordsPointerEvent(t *testing.T) {
	srv := StartFakeVNCServer(t, testImage())
	conn := doHandshake(t, srv.Addr)
	defer conn.Close()

	// Send PointerEvent: msg-type=5, button-mask=1, x=400(0x0190), y=300(0x012c)
	ptrMsg := []byte{5, 1, 0x01, 0x90, 0x01, 0x2c}
	if _, err := conn.Write(ptrMsg); err != nil {
		t.Fatalf("write pointer event error: %v", err)
	}

	// Send release
	ptrMsg2 := []byte{5, 0, 0x01, 0x90, 0x01, 0x2c}
	if _, err := conn.Write(ptrMsg2); err != nil {
		t.Fatalf("write pointer release error: %v", err)
	}

	var events []PointerEvent
	for i := 0; i < 100; i++ {
		events = srv.GetPointerEvents()
		if len(events) >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(events) < 2 {
		t.Fatalf("got %d pointer events, want >= 2", len(events))
	}
	if events[0].X != 400 || events[0].Y != 300 || events[0].ButtonMask != 1 {
		t.Errorf("event[0] = %+v, want {X:400, Y:300, ButtonMask:1}", events[0])
	}
	if events[1].ButtonMask != 0 {
		t.Errorf("event[1].ButtonMask = %d, want 0", events[1].ButtonMask)
	}
}

// doHandshake performs the full RFB 003.008 handshake with SecurityType None.
func doHandshake(t *testing.T, addr string) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}

	// Version exchange
	buf := make([]byte, 12)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read version error: %v", err)
	}
	if _, err := conn.Write([]byte("RFB 003.008\n")); err != nil {
		t.Fatalf("write version error: %v", err)
	}

	// Security types
	secBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, secBuf); err != nil {
		t.Fatalf("read security error: %v", err)
	}
	if _, err := conn.Write([]byte{1}); err != nil {
		t.Fatalf("write security error: %v", err)
	}

	// No SecurityResult for security type None (kward/go-vnc compatibility).

	// ClientInit
	if _, err := conn.Write([]byte{1}); err != nil {
		t.Fatalf("write client init error: %v", err)
	}

	// ServerInit (2+2+16+4 = 24 bytes header + name)
	serverInitBuf := make([]byte, 24)
	if _, err := io.ReadFull(conn, serverInitBuf); err != nil {
		t.Fatalf("read server init error: %v", err)
	}
	nameLen := binary.BigEndian.Uint32(serverInitBuf[20:24])
	if nameLen > 0 {
		nameBuf := make([]byte, nameLen)
		if _, err := io.ReadFull(conn, nameBuf); err != nil {
			t.Fatalf("read desktop name error: %v", err)
		}
	}

	return conn
}
