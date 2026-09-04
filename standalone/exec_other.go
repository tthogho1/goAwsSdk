//go:build !windows

package main

import "os/exec"

// setHideWindow is a no-op on non-Windows platforms since there is no

func setHideWindow(cmd *exec.Cmd) {} // console window to hide.
