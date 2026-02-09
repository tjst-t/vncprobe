package testutil

import (
	"encoding/binary"
	"image"
	"io"
	"net"
	"sync"
	"testing"
)

// KeyEvent records a key event received by the fake server.
type KeyEvent struct {
	Key      uint32
	DownFlag bool
}

// PointerEvent records a pointer event received by the fake server.
type PointerEvent struct {
	X, Y       uint16
	ButtonMask uint8
}

// FakeVNCServer is a minimal RFB 003.008 server for testing.
type FakeVNCServer struct {
	Addr     string
	listener net.Listener
	img      image.Image

	mu        sync.Mutex
	keyEvents []KeyEvent
	ptrEvents []PointerEvent
}

// StartFakeVNCServer starts a fake VNC server on a random port.
// It is automatically stopped via t.Cleanup.
func StartFakeVNCServer(t *testing.T, framebufferImage image.Image) *FakeVNCServer {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}

	srv := &FakeVNCServer{
		Addr:     ln.Addr().String(),
		listener: ln,
		img:      framebufferImage,
	}

	t.Cleanup(func() {
		ln.Close()
	})

	go srv.acceptLoop()

	return srv
}

// GetKeyEvents returns a copy of all recorded key events.
func (s *FakeVNCServer) GetKeyEvents() []KeyEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]KeyEvent, len(s.keyEvents))
	copy(cp, s.keyEvents)
	return cp
}

// GetPointerEvents returns a copy of all recorded pointer events.
func (s *FakeVNCServer) GetPointerEvents() []PointerEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]PointerEvent, len(s.ptrEvents))
	copy(cp, s.ptrEvents)
	return cp
}

func (s *FakeVNCServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return // listener closed
		}
		go s.handleConn(conn)
	}
}

func (s *FakeVNCServer) handleConn(conn net.Conn) {
	defer conn.Close()

	// --- Protocol Version ---
	if _, err := conn.Write([]byte("RFB 003.008\n")); err != nil {
		return
	}

	clientVersion := make([]byte, 12)
	if _, err := io.ReadFull(conn, clientVersion); err != nil {
		return
	}

	// --- Security ---
	// Send: 1 security type, type=None(1)
	if _, err := conn.Write([]byte{1, 1}); err != nil {
		return
	}

	// Read client's selected security type
	secType := make([]byte, 1)
	if _, err := io.ReadFull(conn, secType); err != nil {
		return
	}

	// Send SecurityResult (OK=0)
	if err := binary.Write(conn, binary.BigEndian, uint32(0)); err != nil {
		return
	}

	// --- ClientInit ---
	clientInit := make([]byte, 1)
	if _, err := io.ReadFull(conn, clientInit); err != nil {
		return
	}

	// --- ServerInit ---
	bounds := s.img.Bounds()
	width := uint16(bounds.Dx())
	height := uint16(bounds.Dy())

	// Width, Height
	binary.Write(conn, binary.BigEndian, width)
	binary.Write(conn, binary.BigEndian, height)

	// PixelFormat: 32-bit true color
	pixelFormat := []byte{
		32,      // bits-per-pixel
		24,      // depth
		0,       // big-endian-flag (little)
		1,       // true-color-flag
		0, 255,  // red-max (255)
		0, 255,  // green-max (255)
		0, 255,  // blue-max (255)
		16,      // red-shift
		8,       // green-shift
		0,       // blue-shift
		0, 0, 0, // padding
	}
	conn.Write(pixelFormat)

	// Desktop name
	name := []byte("fake")
	binary.Write(conn, binary.BigEndian, uint32(len(name)))
	conn.Write(name)

	// --- Main message loop ---
	for {
		msgType := make([]byte, 1)
		if _, err := io.ReadFull(conn, msgType); err != nil {
			return
		}

		switch msgType[0] {
		case 0: // SetPixelFormat
			buf := make([]byte, 19) // 3 padding + 16 pixel format
			if _, err := io.ReadFull(conn, buf); err != nil {
				return
			}

		case 2: // SetEncodings
			buf := make([]byte, 3) // 1 padding + 2 num-encodings
			if _, err := io.ReadFull(conn, buf); err != nil {
				return
			}
			numEncodings := binary.BigEndian.Uint16(buf[1:3])
			encBuf := make([]byte, 4*int(numEncodings))
			if _, err := io.ReadFull(conn, encBuf); err != nil {
				return
			}

		case 3: // FramebufferUpdateRequest
			buf := make([]byte, 9) // incremental(1) + x(2) + y(2) + w(2) + h(2)
			if _, err := io.ReadFull(conn, buf); err != nil {
				return
			}
			s.sendFramebufferUpdate(conn)

		case 4: // KeyEvent
			buf := make([]byte, 7) // down-flag(1) + padding(2) + key(4)
			if _, err := io.ReadFull(conn, buf); err != nil {
				return
			}
			down := buf[0] != 0
			key := binary.BigEndian.Uint32(buf[3:7])
			s.mu.Lock()
			s.keyEvents = append(s.keyEvents, KeyEvent{Key: key, DownFlag: down})
			s.mu.Unlock()

		case 5: // PointerEvent
			buf := make([]byte, 5) // button-mask(1) + x(2) + y(2)
			if _, err := io.ReadFull(conn, buf); err != nil {
				return
			}
			mask := buf[0]
			x := binary.BigEndian.Uint16(buf[1:3])
			y := binary.BigEndian.Uint16(buf[3:5])
			s.mu.Lock()
			s.ptrEvents = append(s.ptrEvents, PointerEvent{X: x, Y: y, ButtonMask: mask})
			s.mu.Unlock()

		case 6: // ClientCutText
			buf := make([]byte, 7) // padding(3) + length(4)
			if _, err := io.ReadFull(conn, buf); err != nil {
				return
			}
			textLen := binary.BigEndian.Uint32(buf[3:7])
			textBuf := make([]byte, textLen)
			if _, err := io.ReadFull(conn, textBuf); err != nil {
				return
			}

		default:
			return // unknown message
		}
	}
}

func (s *FakeVNCServer) sendFramebufferUpdate(conn net.Conn) {
	bounds := s.img.Bounds()
	width := uint16(bounds.Dx())
	height := uint16(bounds.Dy())

	// Message type (0) + padding (1) + number-of-rectangles (2)
	header := []byte{0, 0, 0, 1} // 1 rectangle
	conn.Write(header)

	// Rectangle header: x(2) + y(2) + width(2) + height(2) + encoding-type(4)
	binary.Write(conn, binary.BigEndian, uint16(0)) // x
	binary.Write(conn, binary.BigEndian, uint16(0)) // y
	binary.Write(conn, binary.BigEndian, width)
	binary.Write(conn, binary.BigEndian, height)
	binary.Write(conn, binary.BigEndian, int32(0)) // Raw encoding

	// Pixel data: 4 bytes per pixel
	// Our pixel format: 32bpp, red-shift=16, green-shift=8, blue-shift=0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := s.img.At(x, y).RGBA()
			// Pack as uint32: (R<<16) | (G<<8) | B, big-endian on wire
			pixel := uint32(r>>8)<<16 | uint32(g>>8)<<8 | uint32(b>>8)
			binary.Write(conn, binary.BigEndian, pixel)
		}
	}
}
