package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alebak/jkit/internal/agents"
	"github.com/spf13/cobra"
)

var skillsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install skill symlinks for agents",
	Long: `Creates symlinks from agent skill directories to .jkit/agents/skills/.
Run inside a devcontainer or on your host to make skills available to AI agents.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}

		skillName, _ := cmd.Flags().GetString("skill")
		agentName, _ := cmd.Flags().GetString("agent")

		skillDir, err := agents.SkillDirFor(context.Background(), agentName)
		if err != nil {
			return fmt.Errorf("resolving agent dir: %w", err)
		}

		if err := agents.LinkSkill(context.Background(), cwd, skillDir, skillName); err != nil {
			return fmt.Errorf("linking skill: %w", err)
		}

		cmd.Printf("Linked %s → %s skills\n", skillName, agentName)
		return nil
	},
}

func init() {
	skillsCmd := &cobra.Command{
		Use:   "skills",
		Short: "Manage AI skills",
	}
	skillsCmd.AddCommand(skillsInstallCmd)
	rootCmd.AddCommand(skillsCmd)

	skillsInstallCmd.Flags().String("skill", "prd-creator-joomla", "Skill name")
	skillsInstallCmd.Flags().String("agent", "claude", "Agent name (claude, opencode, gemini)")
}
