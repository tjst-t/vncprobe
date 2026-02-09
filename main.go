package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: vncprobe <command> [options]")
		fmt.Fprintln(os.Stderr, "Commands: capture, key, type, click, move")
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "not implemented")
	os.Exit(1)
}
