package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tjst-t/vncprobe/testutil"
)

func e2eImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 128, A: 255})
		}
	}
	return img
}

// runVncprobe calls the run() function directly (same process, no exec).
func runVncprobe(t *testing.T, args ...string) int {
	t.Helper()
	return run(args)
}

func TestE2ECapture(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())
	out := filepath.Join(t.TempDir(), "screen.png")

	code := runVncprobe(t, "capture", "-s", srv.Addr, "-o", out)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("output size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
	}
}

func TestE2EKey(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())

	code := runVncprobe(t, "key", "-s", srv.Addr, "enter")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

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
		t.Errorf("event[0] = %+v, want {Key:0xff0d, DownFlag:true}", events[0])
	}
	if events[1].Key != 0xff0d || events[1].DownFlag != false {
		t.Errorf("event[1] = %+v, want {Key:0xff0d, DownFlag:false}", events[1])
	}
}

func TestE2EKeyCombo(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())

	code := runVncprobe(t, "key", "-s", srv.Addr, "ctrl-c")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	var events []testutil.KeyEvent
	for i := 0; i < 100; i++ {
		events = srv.GetKeyEvents()
		if len(events) >= 4 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(events) < 4 {
		t.Fatalf("got %d key events, want >= 4", len(events))
	}
	if events[0].Key != 0xffe3 || events[0].DownFlag != true {
		t.Errorf("event[0] = %+v, want ctrl press", events[0])
	}
	if events[1].Key != 0x0063 || events[1].DownFlag != true {
		t.Errorf("event[1] = %+v, want 'c' press", events[1])
	}
}

func TestE2EType(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())

	code := runVncprobe(t, "type", "-s", srv.Addr, "Hi")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	var events []testutil.KeyEvent
	for i := 0; i < 100; i++ {
		events = srv.GetKeyEvents()
		if len(events) >= 6 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	// 'H' = Shift+h (4 events), 'i' = normal (2 events) = 6 total
	if len(events) != 6 {
		t.Fatalf("got %d key events, want 6: %+v", len(events), events)
	}
	if events[0].Key != 0xffe1 || events[0].DownFlag != true {
		t.Errorf("event[0] = %+v, want Shift_L press", events[0])
	}
	if events[1].Key != 0x0068 || events[1].DownFlag != true {
		t.Errorf("event[1] = %+v, want 'h' press", events[1])
	}
	if events[2].Key != 0x0068 || events[2].DownFlag != false {
		t.Errorf("event[2] = %+v, want 'h' release", events[2])
	}
	if events[3].Key != 0xffe1 || events[3].DownFlag != false {
		t.Errorf("event[3] = %+v, want Shift_L release", events[3])
	}
	if events[4].Key != 0x0069 || events[4].DownFlag != true {
		t.Errorf("event[4] = %+v, want 'i' press", events[4])
	}
	if events[5].Key != 0x0069 || events[5].DownFlag != false {
		t.Errorf("event[5] = %+v, want 'i' release", events[5])
	}
}

func TestE2ETypeShiftedChars(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())

	code := runVncprobe(t, "type", "-s", srv.Addr, "a!")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	var events []testutil.KeyEvent
	for i := 0; i < 100; i++ {
		events = srv.GetKeyEvents()
		if len(events) >= 6 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	// 'a' = press+release (2 events), '!' = Shift press + '1' press + '1' release + Shift release (4 events)
	if len(events) != 6 {
		t.Fatalf("got %d key events, want 6: %+v", len(events), events)
	}
	// 'a' press
	if events[0].Key != 0x0061 || events[0].DownFlag != true {
		t.Errorf("event[0] = %+v, want 'a' press", events[0])
	}
	// 'a' release
	if events[1].Key != 0x0061 || events[1].DownFlag != false {
		t.Errorf("event[1] = %+v, want 'a' release", events[1])
	}
	// Shift_L press
	if events[2].Key != 0xffe1 || events[2].DownFlag != true {
		t.Errorf("event[2] = %+v, want Shift_L press", events[2])
	}
	// '1' press (base key for '!')
	if events[3].Key != 0x0031 || events[3].DownFlag != true {
		t.Errorf("event[3] = %+v, want '1' press", events[3])
	}
	// '1' release
	if events[4].Key != 0x0031 || events[4].DownFlag != false {
		t.Errorf("event[4] = %+v, want '1' release", events[4])
	}
	// Shift_L release
	if events[5].Key != 0xffe1 || events[5].DownFlag != false {
		t.Errorf("event[5] = %+v, want Shift_L release", events[5])
	}
}

func TestE2EKeyShiftedChar(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())

	code := runVncprobe(t, "key", "-s", srv.Addr, "!")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	var events []testutil.KeyEvent
	for i := 0; i < 100; i++ {
		events = srv.GetKeyEvents()
		if len(events) >= 4 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	// '!' should send: Shift press, '1' press, '1' release, Shift release
	if len(events) != 4 {
		t.Fatalf("got %d key events, want 4: %+v", len(events), events)
	}
	if events[0].Key != 0xffe1 || events[0].DownFlag != true {
		t.Errorf("event[0] = %+v, want Shift_L press", events[0])
	}
	if events[1].Key != 0x0031 || events[1].DownFlag != true {
		t.Errorf("event[1] = %+v, want '1' press", events[1])
	}
	if events[2].Key != 0x0031 || events[2].DownFlag != false {
		t.Errorf("event[2] = %+v, want '1' release", events[2])
	}
	if events[3].Key != 0xffe1 || events[3].DownFlag != false {
		t.Errorf("event[3] = %+v, want Shift_L release", events[3])
	}
}

func TestE2EClick(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())

	code := runVncprobe(t, "click", "-s", srv.Addr, "400", "300")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	var events []testutil.PointerEvent
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
		t.Errorf("press = %+v, want {X:400, Y:300, ButtonMask:1}", events[0])
	}
	if events[1].ButtonMask != 0 {
		t.Errorf("release ButtonMask = %d, want 0", events[1].ButtonMask)
	}
}

func TestE2EClickRightButton(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())

	code := runVncprobe(t, "click", "-s", srv.Addr, "--button", "3", "400", "300")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	var events []testutil.PointerEvent
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
	if events[0].ButtonMask != 4 {
		t.Errorf("press ButtonMask = %d, want 4 (right)", events[0].ButtonMask)
	}
}

func TestE2EMove(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())

	code := runVncprobe(t, "move", "-s", srv.Addr, "500", "600")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
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
	if events[0].X != 500 || events[0].Y != 600 || events[0].ButtonMask != 0 {
		t.Errorf("move = %+v, want {X:500, Y:600, ButtonMask:0}", events[0])
	}
}

func TestE2EWaitChange(t *testing.T) {
	red := solidColorImage(64, 64, color.RGBA{R: 255, A: 255})
	blue := solidColorImage(64, 64, color.RGBA{B: 255, A: 255})

	srv := testutil.StartFakeVNCServer(t, red)

	// Change the image after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		srv.SetImage(blue)
	}()

	code := runVncprobe(t, "wait", "change", "-s", srv.Addr, "--max-wait", "5", "--interval", "0.1")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

func TestE2EWaitChangeTimeout(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())

	// Image never changes → should timeout
	code := runVncprobe(t, "wait", "change", "-s", srv.Addr, "--max-wait", "0.5", "--interval", "0.1")
	if code != 3 {
		t.Fatalf("exit code = %d, want 3 (timeout)", code)
	}
}

func TestE2EWaitStable(t *testing.T) {
	red := solidColorImage(64, 64, color.RGBA{R: 255, A: 255})
	blue := solidColorImage(64, 64, color.RGBA{B: 255, A: 255})

	srv := testutil.StartFakeVNCServer(t, red)

	// Change image once, then stay stable
	go func() {
		time.Sleep(100 * time.Millisecond)
		srv.SetImage(blue)
	}()

	code := runVncprobe(t, "wait", "stable", "-s", srv.Addr, "--duration", "0.5", "--max-wait", "5", "--interval", "0.1")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
}

func TestE2EWaitNoSubcommand(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())
	code := runVncprobe(t, "wait", "-s", srv.Addr)
	if code != 3 {
		t.Fatalf("exit code = %d, want 3", code)
	}
}

func solidColorImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestE2ESessionCapture(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())
	sock := filepath.Join(t.TempDir(), "test.sock")
	out := filepath.Join(t.TempDir(), "screen.png")

	// Start session in background
	done := make(chan int, 1)
	go func() {
		done <- runVncprobe(t, "session", "start", "-s", srv.Addr, "--socket", sock)
	}()

	// Wait for socket to appear
	for i := 0; i < 100; i++ {
		if _, err := os.Stat(sock); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Capture via session
	code := runVncprobe(t, "capture", "--socket", sock, "-o", out)
	if code != 0 {
		t.Fatalf("capture via session: exit code = %d, want 0", code)
	}

	// Verify output file
	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	bounds := decoded.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("output size = %dx%d, want 64x64", bounds.Dx(), bounds.Dy())
	}

	// Stop session
	code = runVncprobe(t, "session", "stop", "--socket", sock)
	if code != 0 {
		t.Fatalf("session stop: exit code = %d, want 0", code)
	}
}

func TestE2ESessionMultipleCommands(t *testing.T) {
	srv := testutil.StartFakeVNCServer(t, e2eImage())
	sock := filepath.Join(t.TempDir(), "test.sock")

	go runVncprobe(t, "session", "start", "-s", srv.Addr, "--socket", sock)

	for i := 0; i < 100; i++ {
		if _, err := os.Stat(sock); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// key via session
	code := runVncprobe(t, "key", "--socket", sock, "enter")
	if code != 0 {
		t.Fatalf("key via session: exit code = %d, want 0", code)
	}

	// type via session
	code = runVncprobe(t, "type", "--socket", sock, "hello")
	if code != 0 {
		t.Fatalf("type via session: exit code = %d, want 0", code)
	}

	runVncprobe(t, "session", "stop", "--socket", sock)
}

func TestE2EMissingServer(t *testing.T) {
	code := runVncprobe(t, "capture")
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (arg error)", code)
	}
}

func TestE2EUnknownCommand(t *testing.T) {
	code := runVncprobe(t, "foobar", "-s", "127.0.0.1:5900")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestE2ENoArgs(t *testing.T) {
	code := runVncprobe(t)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestE2EVersion(t *testing.T) {
	code := runVncprobe(t, "-v")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}

	code = runVncprobe(t, "--version")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}

	code = runVncprobe(t, "version")
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}
