package cmd

import (
	"flag"

	"github.com/tjst-t/vncprobe/vnc"
)

// RunCapture executes the capture command.
func RunCapture(client vnc.VNCClient, args []string) error {
	fs := flag.NewFlagSet("capture", flag.ContinueOnError)
	output := fs.String("o", "screen.png", "Output PNG file path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return vnc.CaptureToFile(client, *output)
}
