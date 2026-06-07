package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

// Executor defines the interface for running a routine.
type Executor interface {
	Run(folder string) (string, error)
	Describe() string
}

type executorType struct {
	Type string `json:"type"`
}

// NewExecutor creates an Executor from raw JSON based on the "type" field.
func NewExecutor(raw json.RawMessage) (Executor, error) {
	var et executorType
	if err := json.Unmarshal(raw, &et); err != nil {
		return nil, fmt.Errorf("cannot read executor type: %w", err)
	}
	switch et.Type {
	case "opencode":
		var e openCodeExecutor
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("invalid opencode executor: %w", err)
		}
		return &e, nil
	case "shell":
		var e shellExecutor
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("invalid shell executor: %w", err)
		}
		return &e, nil
	case "claude":
		var e claudeExecutor
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("invalid claude executor: %w", err)
		}
		return &e, nil
	default:
		return nil, fmt.Errorf("unknown executor type: %q (supported: opencode, shell, claude)", et.Type)
	}
}

// --- Shell Executor ---

type shellExecutor struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

func (e *shellExecutor) Run(folder string) (string, error) {
	cmd := exec.Command("bash", "-c", e.Command)
	cmd.Dir = folder
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (e *shellExecutor) Describe() string {
	if e.Command == "" {
		return "shell"
	}
	desc := e.Command
	if len(desc) > 50 {
		desc = desc[:47] + "..."
	}
	return fmt.Sprintf("shell: %s", desc)
}

// --- OpenCode Executor ---

type openCodeExecutor struct {
	Type   string `json:"type"`
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}

func (e *openCodeExecutor) Run(folder string) (string, error) {
	args := []string{"run", e.Prompt, "--dir", folder, "--dangerously-skip-permissions"}
	if e.Model != "" {
		args = append(args, "--model", e.Model)
	}
	cmd := exec.Command("opencode", args...)
	cmd.Dir = folder
	cmd.Env = append(os.Environ(), "OPENCODE_DISABLE_AUTOUPDATE=true")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (e *openCodeExecutor) Describe() string {
	if e.Prompt == "" {
		return "opencode"
	}
	desc := e.Prompt
	if len(desc) > 50 {
		desc = desc[:47] + "..."
	}
	return fmt.Sprintf("opencode: %s", desc)
}

// --- Claude Executor ---

type claudeExecutor struct {
	Type   string `json:"type"`
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
	// PermissionMode maps to Claude's --permission-mode flag
	// (default, acceptEdits, plan, bypassPermissions).
	PermissionMode string `json:"permissionMode"`
	// DangerouslySkipPermissions is a shorthand for bypassPermissions.
	DangerouslySkipPermissions bool `json:"dangerouslySkipPermissions"`
}

func (e *claudeExecutor) permissionMode() string {
	if e.PermissionMode != "" {
		return e.PermissionMode
	}
	if e.DangerouslySkipPermissions {
		return "bypassPermissions"
	}
	return ""
}

func (e *claudeExecutor) Run(folder string) (string, error) {
	args := []string{"-p", e.Prompt}
	if e.Model != "" {
		args = append(args, "--model", e.Model)
	}
	if mode := e.permissionMode(); mode != "" {
		args = append(args, "--permission-mode", mode)
	}
	cmd := exec.Command("claude", args...)
	cmd.Dir = folder
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (e *claudeExecutor) Describe() string {
	if e.Prompt == "" {
		return "claude"
	}
	desc := e.Prompt
	if len(desc) > 50 {
		desc = desc[:47] + "..."
	}
	return fmt.Sprintf("claude: %s", desc)
}

// Execute runs a routine with logging and notifications.
func Execute(routine Routine) (string, error) {
	executor, err := routine.GetExecutor()
	if err != nil {
		return "", fmt.Errorf("cannot resolve executor: %w", err)
	}

	start := time.Now()
	timestamp := start.Format("2006-01-02 15:04:05")

	execType := routine.ExecutorType()
	desc := executor.Describe()

	AppendLog(fmt.Sprintf("[%s] %-20s START (%s)", timestamp, routine.Name, desc))
	AppendLog(fmt.Sprintf("[%s] %-20s Running in: %s", timestamp, routine.Name, routine.Folder))

	output, err := executor.Run(routine.Folder)

	elapsed := time.Since(start).Round(time.Second)
	if err != nil {
		AppendLog(fmt.Sprintf("[%s] %-20s ERROR (%s): %v", timestamp, routine.Name, elapsed, err))
		if len(output) > 0 {
			AppendLog(fmt.Sprintf("[%s] %-20s OUTPUT:\n%s", timestamp, routine.Name, strings.TrimSpace(output)))
		}
		if routine.Notify {
			beeep.Notify(
				fmt.Sprintf("Routine failed: %s", routine.Name),
				fmt.Sprintf("%s: error after %s", execType, elapsed),
				"",
			)
		}
		return output, fmt.Errorf("execution failed: %w", err)
	}

	AppendLog(fmt.Sprintf("[%s] %-20s DONE (%s)", timestamp, routine.Name, elapsed))
	if routine.Notify {
		beeep.Notify(
			fmt.Sprintf("Routine done: %s", routine.Name),
			fmt.Sprintf("%s completed in %s", execType, elapsed),
			"",
		)
	}
	return output, nil
}
