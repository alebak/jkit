package main

import (
	"context"
	"fmt"
	"os"

	initpkg "github.com/alebak/jkit/internal/init"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Joomla project",
	Long: `Interactive wizard to scaffold a .devcontainer/ configuration
for Joomla development.

Without flags, launches an interactive TUI (requires TTY).
With --name or other flags, runs in parameterized mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Detect if any explicit flags were set
		nameSet := cmd.Flags().Changed("name")
		imageSet := cmd.Flags().Changed("image")
		agentsSet := cmd.Flags().Changed("agents")
		timezoneSet := cmd.Flags().Changed("timezone")
		quickstartSet := cmd.Flags().Changed("quickstart")
		forceSet := cmd.Flags().Changed("force")

		anyFlagSet := nameSet || imageSet || agentsSet || timezoneSet || quickstartSet || forceSet

		if anyFlagSet {
			// Parameterized mode
			return runParameterized(ctx, cmd)
		}

		// Check if stdout is a terminal.
		// os.Stdout.Stat only fails on broken stdout — treat as non-TTY.
		fi, _ := os.Stdout.Stat()
		isTTY := (fi.Mode() & os.ModeCharDevice) != 0

		if !isTTY {
			return fmt.Errorf("run with --name or --image flags in non-TTY mode")
		}

		// Interactive TUI mode
		cfg, err := initpkg.RunInteractive(ctx)
		if err != nil {
			return fmt.Errorf("interactive init cancelled: %w", err)
		}

		if err := initpkg.Orchestrate(ctx, cfg); err != nil {
			return fmt.Errorf("init failed: %w", err)
		}

		cmd.Println("Project initialized successfully.")
		return nil
	},
}

func runParameterized(ctx context.Context, cmd *cobra.Command) error {
	// cobra Flag methods only error when the flag was never registered
	// in init() — all flags below are registered. Errors are unreachable.
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return fmt.Errorf("--name is required in parameterized mode")
	}

	image, _ := cmd.Flags().GetString("image")
	agents, _ := cmd.Flags().GetStringSlice("agents")
	timezone, _ := cmd.Flags().GetString("timezone")
	quickstart, _ := cmd.Flags().GetString("quickstart")
	force, _ := cmd.Flags().GetBool("force")

	cfg := initpkg.DefaultInitConfig()
	cfg.ProjectName = name
	if image != "" {
		cfg.JoomlaImage = image
	}
	cfg.Timezone = timezone
	cfg.Agents = agents
	cfg.Quickstart = quickstart
	cfg.Force = force

	if err := initpkg.Orchestrate(ctx, cfg); err != nil {
		return fmt.Errorf("init failed: %w", err)
	}

	cmd.Println("Project initialized successfully.")
	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("name", "", "Project name")
	initCmd.Flags().String("image", "", "Joomla Docker image tag")
	initCmd.Flags().String("quickstart", "", "Path to quickstart .zip file (auto-detect if empty)")
	initCmd.Flags().StringSlice("agents", nil, "AI agents to install (comma-separated)")
	initCmd.Flags().String("timezone", "UTC", "Timezone for the devcontainer")
	initCmd.Flags().Bool("force", false, "Overwrite existing .devcontainer/ without prompting")
}
