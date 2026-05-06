package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alebak/jkit/internal/agents"
	"github.com/spf13/cobra"
)

const defaultSkillName = "prd-creator"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Joomla project",
	Long: `Interactive wizard to scaffold a .devcontainer/ configuration
for Joomla development.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		agentsList, err := cmd.Flags().GetStringSlice("agents")
		if err != nil {
			return fmt.Errorf("internal: --agents flag not registered: %w", err)
		}

		// Deploy skills for selected agents
		if len(agentsList) > 0 {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot determine current directory: %w", err)
			}
			for _, a := range agentsList {
				skillDir, err := agents.SkillDirFor(context.Background(), a)
				if err != nil {
					cmd.PrintErrf("Warning: unknown agent %s\n", a)
					continue
				}
				if err := agents.DeploySkill(context.Background(), cwd, skillDir, defaultSkillName); err != nil {
					cmd.PrintErrf("Warning: failed to deploy skill for %s: %v\n", a, err)
				}
			}
		}

		cmd.Println("not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("name", "", "Project name")
	initCmd.Flags().String("image", "", "Joomla Docker image tag")
	initCmd.Flags().Bool("quickstart", false, "Generate quickstart config (minimal prompts)")
	initCmd.Flags().StringSlice("agents", nil, "AI agents to install (comma-separated)")
	initCmd.Flags().String("timezone", "UTC", "Timezone for the devcontainer")
}
