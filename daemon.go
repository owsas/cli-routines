package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

func PidPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "routines.pid"), nil
}

func WritePid(pid int) error {
	path, err := PidPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0644)
}

func ReadPid() (int, error) {
	path, err := PidPath()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in %s: %w", path, err)
	}
	return pid, nil
}

func RemovePidFile() {
	path, err := PidPath()
	if err != nil {
		return
	}
	os.Remove(path)
}

func KillDaemon() error {
	pid, err := ReadPid()
	if err != nil {
		return fmt.Errorf("no PID file found; is the daemon running?")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		RemovePidFile()
		return fmt.Errorf("cannot find process %d: %w", pid, err)
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		RemovePidFile()
		return fmt.Errorf("cannot terminate process %d: %w", pid, err)
	}
	RemovePidFile()
	return nil
}
