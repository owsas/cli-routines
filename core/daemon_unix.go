//go:build !windows

package core

import (
	"os"
	"syscall"
)

// IsDaemonRunning checks whether the daemon process is alive by sending
// a null signal, which tests process existence on Unix.
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
