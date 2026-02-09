package main

import (
	"fmt"
	"os"
	"time"

	"github.com/tjst-t/vncprobe/cmd"
	"github.com/tjst-t/vncprobe/vnc"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, cmd.Usage())
		return 1
	}

	command := args[0]
	remaining := args[1:]

	// Validate command before parsing flags or connecting
	switch command {
	case "capture", "key", "type", "click", "move":
		// valid
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprint(os.Stderr, cmd.Usage())
		return 1
	}

	// Parse global flags from remaining args
	opts, cmdArgs, err := cmd.ParseGlobalFlags(remaining)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Connect to VNC server
	client := vnc.NewRealClient()
	if err := client.Connect(opts.Server, opts.Password, time.Duration(opts.Timeout)*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		return 2
	}
	defer client.Close()

	// Dispatch command
	switch command {
	case "capture":
		err = cmd.RunCapture(client, cmdArgs)
	case "key":
		err = cmd.RunKey(client, cmdArgs)
	case "type":
		err = cmd.RunType(client, cmdArgs)
	case "click":
		err = cmd.RunClick(client, cmdArgs)
	case "move":
		err = cmd.RunMove(client, cmdArgs)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 3
	}
	return 0
}
