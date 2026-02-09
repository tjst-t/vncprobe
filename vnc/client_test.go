package vnc

import (
	"image"
	"testing"
	"time"
)

type mockClient struct {
	connected    bool
	closed       bool
	captureImage image.Image
	captureErr   error
	keyEvents    []KeyAction
	ptrEvents    []pointerEvent
}

type pointerEvent struct {
	x, y       uint16
	buttonMask uint8
}

func (m *mockClient) Connect(addr string, password string, timeout time.Duration) error {
	m.connected = true
	return nil
}

func (m *mockClient) Capture() (image.Image, error) {
	return m.captureImage, m.captureErr
}

func (m *mockClient) SendKey(keycode uint32, down bool) error {
	m.keyEvents = append(m.keyEvents, KeyAction{Key: keycode, Down: down})
	return nil
}

func (m *mockClient) SendPointer(x, y uint16, buttonMask uint8) error {
	m.ptrEvents = append(m.ptrEvents, pointerEvent{x: x, y: y, buttonMask: buttonMask})
	return nil
}

func (m *mockClient) Close() error {
	m.closed = true
	return nil
}

func TestMockImplementsInterface(t *testing.T) {
	var _ VNCClient = &mockClient{}
}
