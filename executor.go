package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

func execute(routine Routine) {
	start := time.Now()
	timestamp := start.Format("2006-01-02 15:04:05")
	AppendLog(fmt.Sprintf("[%s] %-20s START", timestamp, routine.Name))

	args := []string{"run", routine.Prompt}
	args = append(args, "--dir", routine.Folder)
	args = append(args, "--dangerously-skip-permissions")
	if routine.Model != "" {
		args = append(args, "--model", routine.Model)
	}

	AppendLog(fmt.Sprintf("[%s] %-20s Running: opencode %s", timestamp, routine.Name, strings.Join(args, " ")))

	cmd := exec.Command("opencode", args...)
	cmd.Env = append(os.Environ(), "OPENCODE_DISABLE_AUTOUPDATE=true")
	output, err := cmd.CombinedOutput()

	elapsed := time.Since(start).Round(time.Second)
	if err != nil {
		AppendLog(fmt.Sprintf("[%s] %-20s ERROR (%s): %v", timestamp, routine.Name, elapsed, err))
		if len(output) > 0 {
			AppendLog(fmt.Sprintf("[%s] %-20s OUTPUT:\n%s", timestamp, routine.Name, string(output)))
		}
		if routine.Notify {
			beeep.Notify(
				fmt.Sprintf("Routine failed: %s", routine.Name),
				fmt.Sprintf("Error after %s: %v", elapsed, err),
				"",
			)
		}
		return
	}

	AppendLog(fmt.Sprintf("[%s] %-20s DONE (%s)", timestamp, routine.Name, elapsed))
	if routine.Notify {
		beeep.Notify(
			fmt.Sprintf("Routine done: %s", routine.Name),
			fmt.Sprintf("Completed in %s", elapsed),
			"",
		)
	}
}
