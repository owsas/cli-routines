package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "routines",
		Short: "Schedule and run local routines",
		Long: `routines lets you define scheduled tasks in a JSON config file.

Each routine defines a schedule (cron), a working folder, and an executor
type — shell commands, OpenCode prompts, or Claude prompts.

Define routines in ~/.cli-routines/routines.json and run them on a cron schedule.`,
		SilenceUsage: true,
	}

	start := &cobra.Command{
		Use:   "start",
		Short: "Start the routines daemon",
		Long:  "Start the scheduler daemon that runs enabled routines on their schedule.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startCmd()
		},
	}
	start.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run in foreground (don't daemonize)")

	stop := &cobra.Command{
		Use:   "stop",
		Short: "Stop the routines daemon",
		Long:  "Stop the running scheduler daemon.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return stopCmd()
		},
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "Show daemon status and next run times",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusCmd()
		},
	}

	list := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all routines",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCmd()
		},
	}

	initCfg := &cobra.Command{
		Use:   "init",
		Short: "Create a default routines config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return initCmd()
		},
	}

	run := &cobra.Command{
		Use:   "run <name>",
		Short: "Run a single routine immediately",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("routine name is required")
			}
			return runCmd(args[0])
		},
	}

	root.AddCommand(start, stop, status, list, initCfg, run)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
