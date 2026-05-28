package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"
)

var (
	foreground bool
)

func initCmd() error {
	cfg := &Config{
		Routines: []Routine{
			{
				Name:        "example-routine",
				Description: "An example routine — edit or remove me",
				Schedule:    "0 9 * * 1-5",
				Folder:      os.Getenv("HOME"),
				ExecutorRaw: json.RawMessage(`{"type": "shell", "command": "echo Hello from cli-routines"}`),
				Enabled:     false,
				Notify:      true,
			},
		},
	}
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	path, _ := ConfigPath()
	fmt.Printf("Created %s\n", path)
	return nil
}

func startCmd() error {
	AppendLog("Daemon starting (foreground mode)")

	if !foreground {
		return startDaemon()
	}

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	enabled := 0
	for _, r := range cfg.Routines {
		if r.Enabled {
			enabled++
		}
	}
	if enabled == 0 {
		return fmt.Errorf("no enabled routines found in config")
	}

	scheduler := NewScheduler()
	if err := scheduler.Start(cfg); err != nil {
		return err
	}

	if err := WritePid(os.Getpid()); err != nil {
		return err
	}
	defer RemovePidFile()

	AppendLog(fmt.Sprintf("Daemon running with PID %d, %d routine(s)", os.Getpid(), enabled))

	fmt.Printf("Started daemon (PID %d) with %d routine(s)\n", os.Getpid(), enabled)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
	scheduler.Stop()
	return nil
}

func startDaemon() error {
	if IsDaemonRunning() {
		pid, _ := ReadPid()
		return fmt.Errorf("daemon already running with PID %d", pid)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	enabled := 0
	for _, r := range cfg.Routines {
		if r.Enabled {
			enabled++
		}
	}
	if enabled == 0 {
		return fmt.Errorf("no enabled routines found in config")
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable: %w", err)
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return fmt.Errorf("cannot resolve executable path: %w", err)
	}

	cmd := exec.Command(exe, "start", "--foreground")
	spawnDaemonProcess(cmd)

	logPath, _ := LogPath()
	if logPath != "" {
		logFile, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if logFile != nil {
			cmd.Stdout = logFile
			cmd.Stderr = logFile
		}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cannot start daemon: %w", err)
	}

	if err := WritePid(cmd.Process.Pid); err != nil {
		return err
	}

	fmt.Printf("Daemon started with PID %d (%d routine(s))\n", cmd.Process.Pid, enabled)
	return nil
}

func stopCmd() error {
	if !IsDaemonRunning() {
		return fmt.Errorf("daemon is not running")
	}
	pid, _ := ReadPid()
	if err := KillDaemon(); err != nil {
		return err
	}
	fmt.Printf("Daemon stopped (PID %d)\n", pid)
	return nil
}

func statusCmd() error {
	if IsDaemonRunning() {
		pid, _ := ReadPid()
		fmt.Printf("Daemon: RUNNING (PID %d)\n\n", pid)
	} else {
		fmt.Println("Daemon: STOPPED")
	}

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Routines) == 0 {
		fmt.Println("No routines configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tEXECUTOR\tENABLED\tNEXT RUN")
	for _, r := range cfg.Routines {
		enabled := "no"
		if r.Enabled {
			enabled = "yes"
		}
		nextRun := "-"
		if r.Enabled {
			s := NewScheduler()
			t, err := s.NextRun(r)
			if err == nil {
				nextRun = t.Format("2006-01-02 15:04")
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, r.ExecutorType(), enabled, nextRun)
	}
	w.Flush()
	return nil
}

func listCmd() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Routines) == 0 {
		fmt.Println("No routines configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tDESCRIPTION\tSCHEDULE\tEXECUTOR\tFOLDER\tENABLED")
	for _, r := range cfg.Routines {
		enabled := "no"
		if r.Enabled {
			enabled = "yes"
		}
		desc := r.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			r.Name, desc, r.Schedule, r.ExecutorType(), r.Folder, enabled)
	}
	w.Flush()
	return nil
}

func runCmd(name string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if IsDaemonRunning() {
		return fmt.Errorf("daemon is already running; stop it first with 'routines stop' before running a single routine")
	}

	var found *Routine
	for i := range cfg.Routines {
		if strings.EqualFold(cfg.Routines[i].Name, name) {
			found = &cfg.Routines[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("no routine named %q", name)
	}

	if err := found.Resolve(); err != nil {
		return fmt.Errorf("cannot resolve executor for %q: %w", name, err)
	}

	fmt.Printf("Running: %s\n  Type: %s\n  Folder: %s\n\n", found.Name, found.ExecutorSummary(), found.Folder)
	start := time.Now()
	execute(*found)
	fmt.Printf("\nFinished in %s\n", time.Since(start).Round(time.Second))
	return nil
}
