package main

import (
	"cli-routines/core"
)

// Routine is re-exported from the core package for CLI commands.
type Routine = core.Routine

// Config is re-exported from the core package.
type Config = core.Config

// Executor is re-exported from the core package.
type Executor = core.Executor

// ConfigDir delegates to core.
func ConfigDir() (string, error) { return core.ConfigDir() }

// ConfigPath delegates to core.
func ConfigPath() (string, error) { return core.ConfigPath() }

// LoadConfig delegates to core.
func LoadConfig() (*Config, error) { return core.LoadConfig() }

// SaveConfig delegates to core.
func SaveConfig(cfg *Config) error { return core.SaveConfig(cfg) }

// LogPath delegates to core.
func LogPath() (string, error) { return core.LogPath() }

// AppendLog delegates to core.
func AppendLog(line string) { core.AppendLog(line) }

// NewScheduler delegates to core.
func NewScheduler() *core.Scheduler { return core.NewScheduler() }

// IsDaemonRunning delegates to core.
func IsDaemonRunning() bool { return core.IsDaemonRunning() }

// ReadPid delegates to core.
func ReadPid() (int, error) { return core.ReadPid() }

// WritePid delegates to core.
func WritePid(pid int) error { return core.WritePid(pid) }

// RemovePidFile delegates to core.
func RemovePidFile() { core.RemovePidFile() }

// KillDaemon delegates to core.
func KillDaemon() error { return core.KillDaemon() }

// execute runs a routine inline (delegates to core).
func execute(routine Routine) { core.Execute(routine) }
