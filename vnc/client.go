package vnc

import (
	"image"
	"time"
)

type VNCClient interface {
	Connect(addr string, password string, timeout time.Duration) error
	Capture() (image.Image, error)
	SendKey(keycode uint32, down bool) error
	SendPointer(x, y uint16, buttonMask uint8) error
	Close() error
}
