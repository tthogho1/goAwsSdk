//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// setHideWindow configures the command to run without popping up a console
// window on Windows.
func setHideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}
