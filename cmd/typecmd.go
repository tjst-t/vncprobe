package cmd

import (
	"fmt"
	"strings"

	"github.com/tjst-t/vncprobe/vnc"
)

// RunType executes the type command.
func RunType(client vnc.VNCClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("type command requires a text argument")
	}

	text := strings.Join(args, " ")
	return vnc.SendTypeString(client, text)
}
