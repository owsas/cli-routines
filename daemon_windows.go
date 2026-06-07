//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

const (
	detachedProcess = 0x00000008
)

// spawnDaemonProcess configures the exec.Cmd to daemonize on Windows
// by detaching the child process from the current console and creating
// a new process group (similar to Unix's Setsid).
func spawnDaemonProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess,
	}
}
