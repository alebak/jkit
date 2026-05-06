package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/alebak/jkit"
	"github.com/alebak/jkit/internal/agents"
	"github.com/alebak/jkit/internal/devcontainer"
	"github.com/spf13/cobra"
)

// postCreatePath is the fixed location of post-create.sh in a JKit
// devcontainer project. It is a variable (not const) for testability.
var postCreatePath = ".devcontainer/post-create.sh"

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage AI agent installations",
	Long: `List, add, or remove AI agents from the devcontainer setup.

Agents are discovered from embedded templates (templates/agents/*.sh).
The post-create.sh file is the authoritative record of installed agents,
using machine-parseable delimiter comments (# --- agent:<name> ---).`,
}

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and installed agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		available, err := agents.ListAvailable(context.Background(), jkit.AgentsFS)
		if err != nil {
			return fmt.Errorf("listing agents: %w", err)
		}

		cmd.Println("Available agents:")
		for _, a := range available {
			cmd.Printf("  %s\n", a)
		}

		cmd.Println()
		cmd.Println("Installed agents (from post-create.sh):")
		if _, err := os.Stat(postCreatePath); err != nil {
			if os.IsNotExist(err) {
				cmd.Println("  (run 'jkit init' to generate post-create.sh)")
			} else {
				return fmt.Errorf("checking post-create.sh: %w", err)
			}
		} else {
			installed, err := agents.ParsePostCreateMarkers(context.Background(), postCreatePath)
			if err != nil {
				return fmt.Errorf("reading post-create.sh: %w", err)
			}
			if len(installed) == 0 {
				cmd.Println("  (none)")
			}
			for _, a := range installed {
				cmd.Printf("  %s\n", a)
			}
		}

		return nil
	},
}

var agentsAddCmd = &cobra.Command{
	Use:   "add [name...]",
	Short: "Add agents to post-create.sh",
	Long: `Add one or more AI agents to the devcontainer post-create.sh script.

Each name is validated against the available agents. The post-create.sh
is regenerated via the renderer with the union of currently installed
and newly requested agents.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		available, err := agents.ListAvailable(context.Background(), jkit.AgentsFS)
		if err != nil {
			return fmt.Errorf("listing agents: %w", err)
		}
		availSet := make(map[string]bool, len(available))
		for _, a := range available {
			availSet[a] = true
		}

		// Validate all requested names
		var invalid []string
		for _, name := range args {
			if !availSet[name] {
				invalid = append(invalid, name)
			}
		}
		if len(invalid) > 0 {
			return fmt.Errorf("unknown agent(s): %s", strings.Join(invalid, ", "))
		}

		if _, err := os.Stat(postCreatePath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("post-create.sh not found at %s; run 'jkit init' first", postCreatePath)
			}
			return fmt.Errorf("checking post-create.sh: %w", err)
		}

		// Parse currently installed agents (add command)
		current, err := agents.ParsePostCreateMarkers(context.Background(), postCreatePath)
		if err != nil {
			return fmt.Errorf("reading post-create.sh: %w", err)
		}

		// Union current + requested
		agentSet := make(map[string]bool, len(current)+len(args))
		for _, a := range current {
			agentSet[a] = true
		}
		for _, a := range args {
			agentSet[a] = true
		}
		newAgents := make([]string, 0, len(agentSet))
		for a := range agentSet {
			newAgents = append(newAgents, a)
		}
		sort.Strings(newAgents)

		// Regenerate post-create.sh via renderer
		data := devcontainer.DevcontainerData{
			SelectedAgents: newAgents,
		}
		var buf bytes.Buffer
		if err := devcontainer.Render(context.Background(), &buf, "post-create.sh", data); err != nil {
			return fmt.Errorf("rendering post-create.sh: %w", err)
		}
		if err := os.WriteFile(postCreatePath, buf.Bytes(), 0755); err != nil {
			return fmt.Errorf("writing post-create.sh: %w", err)
		}

		cmd.Printf("✅ Added %s to post-create.sh\n", strings.Join(args, ", "))
		return nil
	},
}

var agentsRemoveCmd = &cobra.Command{
	Use:   "remove [name...]",
	Short: "Remove agents from post-create.sh",
	Long: `Remove one or more AI agents from the devcontainer post-create.sh script.

The post-create.sh is regenerated with the specified agents removed.
Agents not currently installed produce a warning but do not cause an error.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(postCreatePath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("post-create.sh not found at %s; run 'jkit init' first", postCreatePath)
			}
			return fmt.Errorf("checking post-create.sh: %w", err)
		}

		// Parse currently installed agents (remove command)
		current, err := agents.ParsePostCreateMarkers(context.Background(), postCreatePath)
		if err != nil {
			return fmt.Errorf("reading post-create.sh: %w", err)
		}

		currentSet := make(map[string]bool, len(current))
		for _, a := range current {
			currentSet[a] = true
		}

		// Remove requested agents
		var notInstalled []string
		for _, name := range args {
			if currentSet[name] {
				delete(currentSet, name)
			} else {
				notInstalled = append(notInstalled, name)
			}
		}

		if len(notInstalled) > 0 {
			cmd.Printf("⚠️  Warning: %s not currently installed, skipping\n", strings.Join(notInstalled, ", "))
		}

		newAgents := make([]string, 0, len(currentSet))
		for a := range currentSet {
			newAgents = append(newAgents, a)
		}
		sort.Strings(newAgents)

		// Regenerate post-create.sh via renderer.
		// Use nil SelectedAgents to produce a post-create.sh with no agent sections.
		data := devcontainer.DevcontainerData{
			SelectedAgents: newAgents,
		}
		var buf bytes.Buffer
		if err := devcontainer.Render(context.Background(), &buf, "post-create.sh", data); err != nil {
			return fmt.Errorf("rendering post-create.sh: %w", err)
		}
		if err := os.WriteFile(postCreatePath, buf.Bytes(), 0755); err != nil {
			return fmt.Errorf("writing post-create.sh: %w", err)
		}

		removed := make([]string, 0, len(args))
		for _, a := range args {
			// Check if it was in the original set (i.e., was installed)
			for _, orig := range current {
				if a == orig {
					removed = append(removed, a)
					break
				}
			}
		}
		if len(removed) > 0 {
			cmd.Printf("✅ Removed %s from post-create.sh\n", strings.Join(removed, ", "))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)
	agentsCmd.AddCommand(agentsListCmd)
	agentsCmd.AddCommand(agentsAddCmd)
	agentsCmd.AddCommand(agentsRemoveCmd)
}
