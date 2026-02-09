package cmd

import (
	"fmt"
	"strconv"

	"github.com/tjst-t/vncprobe/vnc"
)

// RunMove executes the move command.
func RunMove(client vnc.VNCClient, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("move command requires x and y coordinates")
	}

	x, err := strconv.ParseUint(args[0], 10, 16)
	if err != nil {
		return fmt.Errorf("invalid x coordinate %q: %w", args[0], err)
	}
	y, err := strconv.ParseUint(args[1], 10, 16)
	if err != nil {
		return fmt.Errorf("invalid y coordinate %q: %w", args[1], err)
	}

	return vnc.SendMove(client, uint16(x), uint16(y))
}
