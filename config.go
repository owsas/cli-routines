package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Routine struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schedule    string          `json:"schedule"`
	Folder      string          `json:"folder"`
	ExecutorRaw json.RawMessage `json:"executor"`
	Enabled     bool            `json:"enabled"`
	Notify      bool            `json:"notify"`

	executor Executor
}

type Config struct {
	Routines []Routine `json:"routines"`
}

func (r *Routine) Resolve() error {
	exec, err := NewExecutor(r.ExecutorRaw)
	if err != nil {
		return err
	}
	r.executor = exec
	return nil
}

func (r *Routine) ExecutorType() string {
	var et struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(r.ExecutorRaw, &et); err != nil {
		return "unknown"
	}
	return et.Type
}

func (r *Routine) ExecutorSummary() string {
	if r.executor == nil {
		return r.ExecutorType()
	}
	return r.executor.Describe()
}

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	dir := filepath.Join(home, ".cli-routines")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create config directory %s: %w", dir, err)
	}
	return dir, nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "routines.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("cannot read config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config %s: %w", path, err)
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cannot write config %s: %w", path, err)
	}
	return nil
}

func LogPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "routines.log"), nil
}

func AppendLog(line string) {
	path, err := LogPath()
	if err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(line + "\n")
}
