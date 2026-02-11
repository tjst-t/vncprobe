package session

// Request represents a command sent to the session server.
type Request struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// Response represents the result of a command execution.
type Response struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}
