package main

import (
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [name]",
	Short: "Build a devcontainer for a project",
	Long: `Generate the .devcontainer/ configuration for a Joomla project
using the specified name or configuration.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
