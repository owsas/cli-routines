# cli-routines

Schedule and run local routines — shell commands, OpenCode prompts, or Claude prompts — on a cron schedule.

## Install

```bash
git clone git@github.com:owsas/cli-routines.git
cd cli-routines
go build -o routines .
sudo cp routines /usr/local/bin/
```

Or with `go install`:

```bash
go install github.com/owsas/cli-routines@latest
```

## Quickstart

```bash
# Create a default config file
routines init

# List your routines
routines list

# Run a routine immediately
routines run example-routine

# Start the daemon (runs enabled routines on schedule)
routines start

# Check daemon status and next run times
routines status

# Stop the daemon
routines stop
```

## Config

Routines are defined in `~/.cli-routines/routines.json`.

### Schema

```json
{
  "routines": [
    {
      "name": "daily-standup",
      "description": "Prep standup notes",
      "schedule": "0 9 * * 1-5",
      "folder": "/Users/juan/my-project",
      "executor": {
        "type": "opencode",
        "prompt": "Summarize git changes from last 24h"
      },
      "enabled": true,
      "notify": true
    }
  ]
}
```

| Field | Type | Description |
|---|---|---|
| `name` | string | Unique identifier for the routine |
| `description` | string | Human-readable description |
| `schedule` | string | Cron expression (minute, hour, day of month, month, day of week) |
| `folder` | string | Working directory when the routine runs |
| `executor` | object | What to run (see executor types below) |
| `enabled` | bool | Whether the daemon should run this on schedule |
| `notify` | bool | Whether to show a desktop notification on completion |

### Executor types (`executor` object)

#### `shell` — Run a shell command

```json
{
  "type": "shell",
  "command": "pg_dump mydb > backup-$(date +%Y%m%d).sql"
}
```

Runs `bash -c "<command>"` in the routine's folder.

#### `opencode` — Run an OpenCode prompt

```json
{
  "type": "opencode",
  "prompt": "Summarize git changes from the last 24 hours",
  "model": ""
}
```

Runs `opencode run "<prompt>" --dir <folder> --dangerously-skip-permissions`.  
Set `model` to a specific model (e.g. `"anthropic/claude-sonnet-4-20250514"`) or leave empty for OpenCode's default.

#### `claude` — Run a Claude prompt

```json
{
  "type": "claude",
  "prompt": "Review this codebase for potential issues",
  "model": "",
  "permissionMode": "",
  "dangerouslySkipPermissions": false
}
```

Runs `claude -p "<prompt>"` in the routine's folder. Requires the [Claude CLI](https://docs.anthropic.com/en/docs/claude-code).  
Set `model` to override the default model.

For **unattended** routines that need to edit files or run tools without an
approval prompt, set the permission mode:

- `permissionMode` — maps to Claude's `--permission-mode` flag. One of
  `default`, `acceptEdits`, `plan`, `bypassPermissions`.
- `dangerouslySkipPermissions` — shorthand for `permissionMode: "bypassPermissions"`
  (mirrors the opencode executor's `--dangerously-skip-permissions`). Ignored
  when `permissionMode` is set explicitly.

Without either, `claude -p` uses whatever mode your Claude settings default to —
which may pause on approval prompts and stall a scheduled run.

### Cron syntax

Standard 5-field cron expressions:

```
┌── minute (0-59)
│ ┌── hour (0-23)
│ │ ┌── day of month (1-31)
│ │ │ ┌── month (1-12)
│ │ │ │ ┌── day of week (0-6, 0=Sunday)
│ │ │ │ │
* * * * *
```

| Expression | Meaning |
|---|---|
| `0 9 * * 1-5` | Every weekday at 9:00 AM |
| `0 14 * * 5` | Every Friday at 2:00 PM |
| `0 3 * * *` | Every day at 3:00 AM |
| `*/15 * * * *` | Every 15 minutes |

Supports `@every` syntax too: `@every 1h30m`, `@every 5m`, `@daily`, `@hourly`.

## CLI Reference

| Command | Description |
|---|---|
| `routines init` | Create `~/.cli-routines/routines.json` with an example routine |
| `routines list` | List all routines (alias: `ls`) |
| `routines status` | Show daemon status and next run times |
| `routines start` | Start the daemon in the background |
| `routines start --foreground` | Start the daemon in the foreground (pinned to terminal) |
| `routines stop` | Stop the running daemon |
| `routines run <name>` | Run a single routine immediately (foreground) |
| `routines tui` | Open the interactive TUI (Bubble Tea terminal dashboard) |

## Terminal UI

`routines tui` opens an interactive terminal UI for managing your routines
without editing JSON by hand. It provides:

- **Dashboard** — See all routines at a glance with daemon status, navigate with
  arrow keys, toggle enabled/disabled, and see next run times.
- **Visual editor** — Add and edit routines with a form interface. Select the
  executor type (shell, opencode, claude) and see only the relevant fields.
- **Run & test** — Select a routine and press `r` to run it immediately with
  live output.
- **Daemon management** — Press `S` to start the daemon, `T` to stop it.
- **Log viewer** — Press `l` to browse execution logs.

### Dashboard keybindings

| Key | Action |
|---|---|
| `↑/↓` or `k/j` | Navigate routines |
| `Enter` or `e` | Edit selected routine |
| `Space` | Toggle enabled |
| `n` | New routine |
| `d` | Delete routine |
| `r` | Run routine |
| `l` | View logs |
| `S` | Start daemon |
| `T` | Stop daemon |
| `Ctrl+R` | Refresh |
| `?` | Help |
| `q` | Quit |

### Editor keybindings

| Key | Action |
|---|---|
| `Tab` / `Shift+Tab` | Navigate fields |
| `Enter` | Toggle bool, cycle select |
| `Ctrl+S` | Save routine |
| `Esc` | Cancel / back |

## Logs

All routine execution is logged to `~/.cli-routines/routines.log`:

```
[2026-05-25 09:00:01] daily-standup        START (opencode: Summarize git changes...)
[2026-05-25 09:00:01] daily-standup        Running in: /Users/juan/my-project
[2026-05-25 09:00:45] daily-standup        DONE (44s)
[2026-05-26 03:00:00] nightly-cleanup      START (shell: find . -name '*.tmp'...)
[2026-05-26 03:00:01] nightly-cleanup      DONE (1s)
```

## Files

| File | Purpose |
|---|---|
| `~/.cli-routines/routines.json` | Routine definitions |
| `~/.cli-routines/routines.pid` | Daemon PID file |
| `~/.cli-routines/routines.log` | Execution log |

## Running on login (macOS)

To have the daemon start automatically when you log in, create a LaunchAgent:

```bash
cat > ~/Library/LaunchAgents/com.cli-routines.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.cli-routines</string>
  <key>ProgramArguments</key>
  <array>
    <string>/usr/local/bin/routines</string>
    <string>start</string>
    <string>--foreground</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>/tmp/routines-launchd.log</string>
  <key>StandardErrorPath</key>
  <string>/tmp/routines-launchd.log</string>
</dict>
</plist>
EOF

launchctl load ~/Library/LaunchAgents/com.cli-routines.plist
```
