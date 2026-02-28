# CLAUDE.md

## Project

vncprobe — Go CLI tool for VNC screen capture, keyboard input, and mouse operations.

## Build & Test

```bash
go build -o vncprobe .
go test ./...
go vet ./...
```

All tests run offline using a fake VNC server in `testutil/`. No external VNC server or network required.

## Architecture

- `main.go` — Entry point. Parses subcommand, global flags, connects to VNC, dispatches to `cmd/`.
- `cmd/` — One file per subcommand (capture, key, typecmd, click, move, wait, session). `root.go` has global flag parsing.
- `vnc/client.go` — `VNCClient` interface. All commands use this interface, not the concrete client.
- `vnc/realclient.go` — Implements `VNCClient` using `github.com/kward/go-vnc` (via fork `tjst-t/go-vnc` with SecurityResult fix).
- `vnc/keymap.go` — `ParseKeySequence("ctrl-c")` returns press/release actions with X11 keysym codes.
- `vnc/input.go` — `SendKeySequence`, `SendTypeString`, `SendClick`, `SendMove` helpers.
- `vnc/capture.go` — `CaptureToFile` captures framebuffer and writes PNG.
- `vnc/compare.go` — `DiffRatio` for pixel-level image comparison.
- `vnc/wait.go` — `WaitForChange`, `WaitForStable` polling loops.
- `session/` — Session server (UNIX socket) and client for persistent VNC connections.
- `testutil/fakeserver.go` — Minimal RFB 003.008 server for testing. Records key/pointer events. `SetImage()` for dynamic screen changes.

## Key conventions

- Exit codes: 0=success, 1=arg error, 2=connection error, 3=operation error.
- Global flags (`-s`, `-p`, `--timeout`, `--socket`) are parsed manually (not `flag.FlagSet`) to allow mixing with subcommand flags.
- `--socket` enables session mode; when set, `-s` is not required.
- `wait` subcommands use `--max-wait` (not `--timeout`) to avoid collision with the global connection timeout.
- Button numbers in CLI (1=left, 2=middle, 3=right) are converted to RFB bitmasks (1, 2, 4) by `ButtonNumberToMask`.

## kward/go-vnc gotchas

- Uses fork `tjst-t/go-vnc` via `replace` directive in go.mod (fixes SecurityResult for SecurityType None).
- `NewClientConfig()` does NOT allocate `ServerMessageCh` — you must create it.
- `ListenAndHandle()` calls `defer Close()` — connection closes when it returns.
- `SetSettle(0)` disables the 25ms UI settle delay (important for tests).
- Only `RawEncoding` is implemented. `Color.R/G/B` are `uint16` — cast to `uint8`.

## Testing

- Unit tests: `vnc/` package uses a mock client (`client_test.go`).
- Integration tests: `vnc/realclient_test.go` tests the real client against the fake server.
- E2E tests: `e2e_test.go` calls `run()` directly with fake server — full CLI flow.
- Fake server: `testutil/fakeserver.go` uses port `:0` (OS auto-assign) to avoid conflicts.
- Fake server sends `SecurityResult` for all `SecurityType`s (RFB 3.8 compliant).
- Fake server must handle `SetPixelFormat` dynamically — client may change pixel format after connect.

## Go Environment

- Go is installed at `/usr/local/go/bin/go` — may not be in default `PATH`. If `go` is not found, prefix with: `export PATH=$PATH:/usr/local/go/bin`
