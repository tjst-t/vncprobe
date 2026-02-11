package vnc

import (
	"errors"
	"fmt"
	"time"
)

// ErrTimeout is returned when a wait operation exceeds its timeout.
var ErrTimeout = errors.New("timeout")

// IsTimeout reports whether err is a timeout error.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// WaitOptions configures wait operations.
type WaitOptions struct {
	Timeout   time.Duration
	Interval  time.Duration
	Threshold float64
}

// WaitForChange captures repeatedly until the screen differs from the initial capture.
func WaitForChange(client VNCClient, opts WaitOptions) error {
	base, err := client.Capture()
	if err != nil {
		return fmt.Errorf("initial capture: %w", err)
	}

	deadline := time.After(opts.Timeout)
	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return fmt.Errorf("screen did not change within %v: %w", opts.Timeout, ErrTimeout)
		case <-ticker.C:
			current, err := client.Capture()
			if err != nil {
				return fmt.Errorf("capture: %w", err)
			}
			ratio, err := DiffRatio(base, current)
			if err != nil {
				return fmt.Errorf("compare: %w", err)
			}
			if ratio > opts.Threshold {
				return nil
			}
		}
	}
}

// WaitForStable captures repeatedly until the screen stays unchanged for stableDuration.
func WaitForStable(client VNCClient, opts WaitOptions, stableDuration time.Duration) error {
	prev, err := client.Capture()
	if err != nil {
		return fmt.Errorf("initial capture: %w", err)
	}

	deadline := time.After(opts.Timeout)
	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	stableSince := time.Now()

	for {
		select {
		case <-deadline:
			return fmt.Errorf("screen did not stabilize within %v: %w", opts.Timeout, ErrTimeout)
		case <-ticker.C:
			current, err := client.Capture()
			if err != nil {
				return fmt.Errorf("capture: %w", err)
			}
			ratio, err := DiffRatio(prev, current)
			if err != nil {
				return fmt.Errorf("compare: %w", err)
			}
			if ratio > opts.Threshold {
				stableSince = time.Now()
			}
			prev = current
			if time.Since(stableSince) >= stableDuration {
				return nil
			}
		}
	}
}
