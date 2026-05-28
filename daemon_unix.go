//go:build !windows

package main

import (
	"os"
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

// IsDaemonRunning checks whether a daemon process is alive by sending
// signal 0 (null signal), which tests process existence on Unix.
func IsDaemonRunning() bool {
	pid, err := ReadPid()
	if err != nil {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		RemovePidFile()
		return false
	}
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		RemovePidFile()
		return false
	}
	return true
}


