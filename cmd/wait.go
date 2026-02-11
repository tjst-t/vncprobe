package cmd

import (
	"flag"
	"fmt"
	"time"

	"github.com/tjst-t/vncprobe/vnc"
)

// RunWait executes the wait command (change or stable subcommand).
func RunWait(client vnc.VNCClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("wait requires a subcommand: change, stable")
	}

	subcmd := args[0]
	subArgs := args[1:]

	switch subcmd {
	case "change":
		return runWaitChange(client, subArgs)
	case "stable":
		return runWaitStable(client, subArgs)
	default:
		return fmt.Errorf("unknown wait subcommand: %s (expected: change, stable)", subcmd)
	}
}

func parseWaitFlags(fs *flag.FlagSet) (timeout, interval, threshold *float64) {
	timeout = fs.Float64("max-wait", 30, "Maximum wait time in seconds")
	interval = fs.Float64("interval", 1, "Polling interval in seconds")
	threshold = fs.Float64("threshold", 0.01, "Pixel difference ratio threshold (0.0-1.0)")
	return
}

func runWaitChange(client vnc.VNCClient, args []string) error {
	fs := flag.NewFlagSet("wait change", flag.ContinueOnError)
	timeout, interval, threshold := parseWaitFlags(fs)

	if err := fs.Parse(args); err != nil {
		return err
	}

	opts := vnc.WaitOptions{
		Timeout:   time.Duration(*timeout * float64(time.Second)),
		Interval:  time.Duration(*interval * float64(time.Second)),
		Threshold: *threshold,
	}
	return vnc.WaitForChange(client, opts)
}

func runWaitStable(client vnc.VNCClient, args []string) error {
	fs := flag.NewFlagSet("wait stable", flag.ContinueOnError)
	timeout, interval, threshold := parseWaitFlags(fs)
	duration := fs.Float64("duration", 0, "Required stable duration in seconds (required)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *duration <= 0 {
		return fmt.Errorf("--duration is required and must be > 0")
	}

	opts := vnc.WaitOptions{
		Timeout:   time.Duration(*timeout * float64(time.Second)),
		Interval:  time.Duration(*interval * float64(time.Second)),
		Threshold: *threshold,
	}
	return vnc.WaitForStable(client, opts, time.Duration(*duration*float64(time.Second)))
}
