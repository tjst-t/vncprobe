package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/tjst-t/vncprobe/cmd"
	"github.com/tjst-t/vncprobe/vnc"
)

// Server maintains a VNC connection and accepts commands over a UNIX socket.
type Server struct {
	client      vnc.VNCClient
	socketPath  string
	idleTimeout time.Duration
	listener    net.Listener
	mu          sync.Mutex
	stopCh      chan struct{}
}

// NewServer creates a new session server.
// If idleTimeout is 0, the server will not auto-shutdown.
func NewServer(client vnc.VNCClient, socketPath string, idleTimeout time.Duration) *Server {
	return &Server{
		client:      client,
		socketPath:  socketPath,
		idleTimeout: idleTimeout,
		stopCh:      make(chan struct{}),
	}
}

// ListenAndServe starts listening on the UNIX socket and serving commands.
// It blocks until Shutdown is called or idle timeout is reached.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.socketPath, err)
	}
	s.listener = ln

	defer func() {
		ln.Close()
		os.Remove(s.socketPath)
	}()

	// Accept connections in a goroutine
	connCh := make(chan net.Conn)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			connCh <- conn
		}
	}()

	var idleTimer <-chan time.Time
	if s.idleTimeout > 0 {
		t := time.NewTimer(s.idleTimeout)
		defer t.Stop()
		idleTimer = t.C
	}

	for {
		select {
		case <-s.stopCh:
			return nil
		case conn := <-connCh:
			s.handleConn(conn)
			// Reset idle timer
			if s.idleTimeout > 0 {
				t := time.NewTimer(s.idleTimeout)
				defer t.Stop()
				idleTimer = t.C
			}
		case <-idleTimer:
			return nil
		}
	}
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() error {
	select {
	case <-s.stopCh:
		// already stopped
	default:
		close(s.stopCh)
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	s.mu.Lock()
	defer s.mu.Unlock()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			writeResponse(conn, Response{OK: false, Error: fmt.Sprintf("invalid request: %v", err)})
			return
		}

		if req.Command == "session" && len(req.Args) > 0 && req.Args[0] == "stop" {
			writeResponse(conn, Response{OK: true})
			go s.Shutdown()
			return
		}

		err := s.dispatchCommand(req.Command, req.Args)
		if err != nil {
			writeResponse(conn, Response{OK: false, Error: err.Error()})
		} else {
			writeResponse(conn, Response{OK: true})
		}
	}
}

func (s *Server) dispatchCommand(command string, args []string) error {
	switch command {
	case "capture":
		return cmd.RunCapture(s.client, args)
	case "key":
		return cmd.RunKey(s.client, args)
	case "type":
		return cmd.RunType(s.client, args)
	case "click":
		return cmd.RunClick(s.client, args)
	case "move":
		return cmd.RunMove(s.client, args)
	case "wait":
		return cmd.RunWait(s.client, args)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func writeResponse(conn net.Conn, resp Response) {
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	conn.Write(data)
}
