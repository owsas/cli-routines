//go:build windows

package core

import (
	"syscall"
)

const (
	stillActive = 0x00000103
)

// IsDaemonRunning checks whether the daemon process is alive on Windows.
// os.FindProcess always succeeds on Windows, so we use OpenProcess
// and GetExitCodeProcess to verify the process is still running.
func IsDaemonRunning() bool {
	pid, err := ReadPid()
	if err != nil {
		return false
	}

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
