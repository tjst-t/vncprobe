package main

import (
	"fmt"
	"os"
	"time"

	"github.com/tjst-t/vncprobe/cmd"
	"github.com/tjst-t/vncprobe/session"
	"github.com/tjst-t/vncprobe/vnc"
)

var version = "dev"

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
	case "-v", "--version", "version":
		fmt.Println(version)
		return 0
	case "session":
		return runSession(remaining)
	case "capture", "key", "type", "click", "move", "wait":
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

	// If --socket is set, route through session client
	if opts.Socket != "" {
		return runViaSession(opts.Socket, command, cmdArgs)
	}

	// Connect to VNC server directly
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
	case "wait":
		err = cmd.RunWait(client, cmdArgs)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 3
	}
	return 0
}

func runSession(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: vncprobe session <start|stop> [options]")
		return 1
	}

	subcmd := args[0]
	subArgs := args[1:]

	switch subcmd {
	case "start":
		opts, err := cmd.ParseSessionStart(subArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		client := vnc.NewRealClient()
		if err := client.Connect(opts.Server, opts.Password, time.Duration(opts.Timeout)*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
			return 2
		}
		idleTimeout := time.Duration(opts.IdleTimeout) * time.Second
		srv := session.NewServer(client, opts.SocketPath, idleTimeout)
		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "Session error: %v\n", err)
			client.Close()
			return 3
		}
		client.Close()
		return 0

	case "stop":
		socketPath, err := cmd.ParseSessionStop(subArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		c := session.NewClient(socketPath)
		if err := c.Execute("session", []string{"stop"}); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 3
		}
		return 0

	default:
		fmt.Fprintf(os.Stderr, "Unknown session subcommand: %s\n", subcmd)
		return 1
	}
}

func runViaSession(socketPath string, command string, args []string) int {
	c := session.NewClient(socketPath)
	if err := c.Execute(command, args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 3
	}
	return 0
}
