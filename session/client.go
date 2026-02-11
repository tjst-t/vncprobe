package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

// Client connects to a session server over a UNIX socket.
type Client struct {
	socketPath string
}

// NewClient creates a new session client.
func NewClient(socketPath string) *Client {
	return &Client{socketPath: socketPath}
}

// Execute sends a command to the session server and returns the result.
func (c *Client) Execute(command string, args []string) error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("connect to session: %w", err)
	}
	defer conn.Close()

	req := Request{Command: command, Args: args}
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("read response: %w", err)
		}
		return fmt.Errorf("no response from session")
	}

	var resp Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}
