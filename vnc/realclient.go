package vnc

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net"
	"time"

	govnc "github.com/kward/go-vnc"
	"github.com/kward/go-vnc/buttons"
	"github.com/kward/go-vnc/keys"
	"github.com/kward/go-vnc/messages"
	"github.com/kward/go-vnc/rfbflags"
)

// Verify that RealClient implements VNCClient at compile time.
var _ VNCClient = (*RealClient)(nil)

// RealClient implements VNCClient using github.com/kward/go-vnc.
type RealClient struct {
	conn   *govnc.ClientConn
	nc     net.Conn
	config *govnc.ClientConfig
	msgCh  chan govnc.ServerMessage
}

// NewRealClient creates a new RealClient.
func NewRealClient() *RealClient {
	return &RealClient{}
}

func (c *RealClient) Connect(addr string, password string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dialer := net.Dialer{Timeout: timeout}
	nc, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", addr, err)
	}

	cfg := govnc.NewClientConfig(password)
	c.msgCh = make(chan govnc.ServerMessage, 100)
	cfg.ServerMessageCh = c.msgCh

	govnc.SetSettle(0) // disable UI settle delay for automation

	vc, err := govnc.Connect(ctx, nc, cfg)
	if err != nil {
		nc.Close()
		return fmt.Errorf("VNC handshake with %s: %w", addr, err)
	}

	c.conn = vc
	c.nc = nc
	c.config = cfg

	// Start listening for server messages in background
	go vc.ListenAndHandle()

	return nil
}

func (c *RealClient) Capture() (image.Image, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	w := c.conn.FramebufferWidth()
	h := c.conn.FramebufferHeight()

	// Request full framebuffer update
	if err := c.conn.FramebufferUpdateRequest(rfbflags.RFBFalse, 0, 0, w, h); err != nil {
		return nil, fmt.Errorf("framebuffer update request: %w", err)
	}

	// Wait for FramebufferUpdate response
	timeout := time.After(10 * time.Second)
	for {
		select {
		case msg := <-c.msgCh:
			if msg.Type() == messages.FramebufferUpdate {
				fbu, ok := msg.(*govnc.FramebufferUpdate)
				if !ok {
					return nil, fmt.Errorf("unexpected message type for FramebufferUpdate")
				}
				return framebufferToImage(w, h, fbu), nil
			}
			// Discard non-framebuffer messages
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for framebuffer update")
		}
	}
}

func (c *RealClient) SendKey(keycode uint32, down bool) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.KeyEvent(keys.Key(keycode), down)
}

func (c *RealClient) SendPointer(x, y uint16, buttonMask uint8) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.PointerEvent(buttons.Button(buttonMask), x, y)
}

func (c *RealClient) Close() error {
	if c.conn != nil {
		log.SetOutput(io.Discard)
		return c.conn.Close()
	}
	return nil
}

// framebufferToImage converts a FramebufferUpdate to an image.RGBA.
func framebufferToImage(width, height uint16, fbu *govnc.FramebufferUpdate) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	for _, rect := range fbu.Rects {
		raw, ok := rect.Enc.(*govnc.RawEncoding)
		if !ok {
			continue
		}
		i := 0
		for y := int(rect.Y); y < int(rect.Y+rect.Height); y++ {
			for x := int(rect.X); x < int(rect.X+rect.Width); x++ {
				if i < len(raw.Colors) {
					clr := raw.Colors[i]
					img.Set(x, y, color.RGBA{
						R: uint8(clr.R),
						G: uint8(clr.G),
						B: uint8(clr.B),
						A: 255,
					})
					i++
				}
			}
		}
	}

	return img
}
