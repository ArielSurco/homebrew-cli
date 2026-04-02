package project

import "github.com/spf13/cobra"

// Cmd is the "project" subcommand that groups project management commands.
var Cmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(removeCmd)
	Cmd.AddCommand(goCmd)
	Cmd.AddCommand(devCmd)
}
