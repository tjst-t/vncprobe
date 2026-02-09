package cmd

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/tjst-t/vncprobe/vnc"
)

// RunClick executes the click command.
func RunClick(client vnc.VNCClient, args []string) error {
	fs := flag.NewFlagSet("click", flag.ContinueOnError)
	button := fs.Int("button", 1, "Mouse button (1=left, 2=middle, 3=right)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	remaining := fs.Args()
	if len(remaining) < 2 {
		return fmt.Errorf("click command requires x and y coordinates")
	}

	x, err := strconv.ParseUint(remaining[0], 10, 16)
	if err != nil {
		return fmt.Errorf("invalid x coordinate %q: %w", remaining[0], err)
	}
	y, err := strconv.ParseUint(remaining[1], 10, 16)
	if err != nil {
		return fmt.Errorf("invalid y coordinate %q: %w", remaining[1], err)
	}

	mask := ButtonNumberToMask(*button)
	return vnc.SendClick(client, uint16(x), uint16(y), mask)
}
