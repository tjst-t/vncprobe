package cmd

import (
	"fmt"
	"strings"
)

// GlobalOpts holds the global CLI options.
type GlobalOpts struct {
	Server   string
	Password string
	Timeout  int
	Socket   string
}

// globalStringFlags maps flag names that take a string value.
var globalStringFlags = map[string]bool{
	"-s": true, "--server": true,
	"-p": true, "--password": true,
	"--socket": true,
}

// globalIntFlags maps flag names that take an int value.
var globalIntFlags = map[string]bool{
	"--timeout": true,
}

// ParseGlobalFlags extracts known global flags from args and returns remaining
// args that belong to the subcommand. Unknown flags are passed through.
func ParseGlobalFlags(args []string) (*GlobalOpts, []string, error) {
	opts := &GlobalOpts{Timeout: 10}
	var remaining []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if globalStringFlags[arg] {
			if i+1 >= len(args) {
				return nil, nil, fmt.Errorf("flag %s requires an argument", arg)
			}
			i++
			val := args[i]
			switch arg {
			case "-s", "--server":
				opts.Server = val
			case "-p", "--password":
				opts.Password = val
			case "--socket":
				opts.Socket = val
			}
		} else if globalIntFlags[arg] {
			if i+1 >= len(args) {
				return nil, nil, fmt.Errorf("flag %s requires an argument", arg)
			}
			i++
			val := args[i]
			switch arg {
			case "--timeout":
				n, err := fmt.Sscanf(val, "%d", &opts.Timeout)
				if n != 1 || err != nil {
					return nil, nil, fmt.Errorf("invalid timeout value: %s", val)
				}
			}
		} else {
			remaining = append(remaining, arg)
		}
	}

	if opts.Server == "" && opts.Socket == "" {
		return nil, nil, fmt.Errorf("server address is required (-s or --server)")
	}

	return opts, remaining, nil
}

// ButtonNumberToMask converts a user-friendly button number (1=left, 2=middle, 3=right)
// to the RFB button bitmask.
func ButtonNumberToMask(button int) uint8 {
	switch button {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 4
	default:
		return 1
	}
}

// Usage prints the usage message.
func Usage() string {
	var b strings.Builder
	b.WriteString("Usage: vncprobe <command> [options]\n\n")
	b.WriteString("Commands:\n")
	b.WriteString("  capture   Capture screen to PNG\n")
	b.WriteString("  key       Send key input\n")
	b.WriteString("  type      Type a string\n")
	b.WriteString("  click     Mouse click\n")
	b.WriteString("  move      Mouse move\n")
	b.WriteString("  wait      Wait for screen change or stability\n")
	b.WriteString("  session   Manage persistent VNC sessions\n")
	b.WriteString("\nGlobal Options:\n")
	b.WriteString("  -s, --server    VNC server address (required)\n")
	b.WriteString("  -p, --password  VNC password\n")
	b.WriteString("  --timeout       Connection timeout in seconds (default: 10)\n")
	b.WriteString("  --socket        Use session socket instead of direct connection\n")
	return b.String()
}
