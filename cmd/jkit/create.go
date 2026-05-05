package main

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [type]",
	Short: "Scaffold a new Joomla extension",
	Long: `Scaffold a new Joomla extension of the specified type.
Supported types: component, module, plugin, template, library, package.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("not yet implemented")
	},
	ValidArgs: []string{
		"component",
		"module",
		"plugin",
		"template",
		"library",
		"package",
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
