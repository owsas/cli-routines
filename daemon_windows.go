//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

const (
	detachedProcess    = 0x00000008
	stillActive        = 0x00000103
)

// spawnDaemonProcess configures the exec.Cmd to daemonize on Windows
// by detaching the child process from the current console and creating
// a new process group (similar to Unix's Setsid).
func spawnDaemonProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess,
	}
}

// IsDaemonRunning checks whether a daemon process is alive on Windows.
// os.FindProcess always succeeds on Windows, so we use OpenProcess
// and GetExitCodeProcess to verify the process is still running.
func IsDaemonRunning() bool {
	pid, err := ReadPid()
	if err != nil {
		return false
	}

	// Open the process with query information access to check its status.
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		RemovePidFile()
		return false
	}
	defer syscall.CloseHandle(handle)

	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	if err != nil || exitCode != stillActive {
		RemovePidFile()
		return false
	}
	return true
}
