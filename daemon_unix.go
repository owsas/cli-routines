//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// spawnDaemonProcess configures the exec.Cmd to daemonize on Unix systems
// by creating a new session (detaching from the parent's terminal).
func spawnDaemonProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}
