package main

import (
	"context"
	"fmt"

	"github.com/alebak/jkit"
	"github.com/alebak/jkit/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP server installations",
	Long: `List, add, or remove MCP server configurations for AI agents.

MCP servers are discovered from embedded templates (templates/mcp/*.json)
and deployed to per-agent mcp.json config files.`,
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available MCP servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		available, err := mcp.ListAvailable(context.Background(), jkit.McpFS)
		if err != nil {
			return fmt.Errorf("listing MCP servers: %w", err)
		}

		if len(available) == 0 {
			cmd.Println("No MCP servers available.")
			return nil
		}

		cmd.Println("Available MCP servers:")
		for _, a := range available {
			cmd.Printf("  %s\n", a)
		}
		return nil
	},
}

var mcpAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add an MCP server to an agent config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		available, err := mcp.ListAvailable(context.Background(), jkit.McpFS)
		if err != nil {
			return fmt.Errorf("listing MCP servers: %w", err)
		}
		availSet := make(map[string]bool, len(available))
		for _, a := range available {
			availSet[a] = true
		}
		if !availSet[name] {
			return fmt.Errorf("unknown MCP: %s", name)
		}

		agent, err := cmd.Flags().GetString("agent")
		if err != nil {
			return fmt.Errorf("internal: --agent flag not registered: %w", err)
		}
		configPath, err := mcp.MCPConfigPathFor(context.Background(), agent)
		if err != nil {
			return fmt.Errorf("resolving config path: %w", err)
		}

		templateData, err := jkit.McpFS.ReadFile("templates/mcp/" + name + ".json")
		if err != nil {
			return fmt.Errorf("reading template: %w", err)
		}

		if err := mcp.DeployMCP(context.Background(), configPath, name, templateData); err != nil {
			return fmt.Errorf("deploying MCP: %w", err)
		}

		cmd.Printf("Added %s to %s config\n", name, agent)
		return nil
	},
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an MCP server from an agent config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		agent, err := cmd.Flags().GetString("agent")
		if err != nil {
			return fmt.Errorf("internal: --agent flag not registered: %w", err)
		}
		configPath, err := mcp.MCPConfigPathFor(context.Background(), agent)
		if err != nil {
			return fmt.Errorf("resolving config path: %w", err)
		}

		if err := mcp.RemoveMCP(context.Background(), configPath, name); err != nil {
			return fmt.Errorf("removing MCP: %w", err)
		}

		cmd.Printf("Removed %s from %s config\n", name, agent)
		return nil
	},
}

func init() {
	mcpCmd.AddCommand(mcpListCmd)
	mcpCmd.AddCommand(mcpAddCmd)
	mcpCmd.AddCommand(mcpRemoveCmd)
	rootCmd.AddCommand(mcpCmd)

	mcpAddCmd.Flags().String("agent", "opencode", "Target AI agent (claude, opencode, gemini)")
	mcpRemoveCmd.Flags().String("agent", "opencode", "Target AI agent (claude, opencode, gemini)")
}
