package main

import (
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Joomla project",
	Long: `Interactive wizard to scaffold a .devcontainer/ configuration
for Joomla development.`,
	Run: func(cmd *cobra.Command, args []string) {
		// cobra Flag methods only error when the flag was never
		// registered — all five flags are registered in init() below.
		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")
		quickstart, _ := cmd.Flags().GetBool("quickstart")
		agents, _ := cmd.Flags().GetStringSlice("agents")
		timezone, _ := cmd.Flags().GetString("timezone")

		_ = name
		_ = image
		_ = quickstart
		_ = agents
		_ = timezone

		cmd.Println("not yet implemented")
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
