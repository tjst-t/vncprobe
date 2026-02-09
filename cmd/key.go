package cmd

import (
	"fmt"

	"github.com/tjst-t/vncprobe/vnc"
)

// RunKey executes the key command.
func RunKey(client vnc.VNCClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("key command requires a key name argument")
	}

	keyStr := args[0]
	actions, err := vnc.ParseKeySequence(keyStr)
	if err != nil {
		return fmt.Errorf("parse key %q: %w", keyStr, err)
	}

	return vnc.SendKeySequence(client, actions)
}
