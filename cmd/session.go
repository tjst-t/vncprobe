package cmd

import (
	"flag"
	"fmt"
)

// SessionStartOpts holds the options for session start.
type SessionStartOpts struct {
	Server      string
	Password    string
	Timeout     int
	SocketPath  string
	IdleTimeout int
}

// ParseSessionStart parses the session start arguments.
func ParseSessionStart(args []string) (*SessionStartOpts, error) {
	fs := flag.NewFlagSet("session start", flag.ContinueOnError)
	server := fs.String("s", "", "VNC server address")
	password := fs.String("p", "", "VNC password")
	timeout := fs.Int("timeout", 10, "Connection timeout in seconds")
	socketPath := fs.String("socket", "", "UNIX socket path")
	idleTimeout := fs.Int("idle-timeout", 300, "Idle timeout in seconds (0 to disable)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if *server == "" {
		return nil, fmt.Errorf("server address is required (-s)")
	}
	if *socketPath == "" {
		return nil, fmt.Errorf("socket path is required (--socket)")
	}

	return &SessionStartOpts{
		Server:      *server,
		Password:    *password,
		Timeout:     *timeout,
		SocketPath:  *socketPath,
		IdleTimeout: *idleTimeout,
	}, nil
}

// ParseSessionStop parses the session stop arguments and returns the socket path.
func ParseSessionStop(args []string) (string, error) {
	fs := flag.NewFlagSet("session stop", flag.ContinueOnError)
	socketPath := fs.String("socket", "", "UNIX socket path")

	if err := fs.Parse(args); err != nil {
		return "", err
	}

	if *socketPath == "" {
		return "", fmt.Errorf("socket path is required (--socket)")
	}

	return *socketPath, nil
}
