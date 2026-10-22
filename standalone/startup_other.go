//go:build !windows

package main

import (
	"fmt"
	"os"
)

// showStartupError prints the startup error to stderr on non-Windows platforms.
func showStartupError(title, msg string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", title, msg)
}
