package main

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command for the jkit CLI.
// It is a package-level var because cobra requires subcommands
// registered via init() in sibling files to access it through
// rootCmd.AddCommand(). This is an accepted exception to the
// "no global mutable state" rule — documented for reviewers.
var rootCmd = &cobra.Command{
	Use:   "jkit",
	Short: "JKit — Joomla project scaffolding toolkit",
	Long: `JKit generates devcontainer configurations, extension scaffolds,
and project templates for Joomla development.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
