package cmd

import (
	cmdproject "github.com/ArielSurco/cli/cmd/project"
	"github.com/spf13/cobra"
)

// Version is injected at build time.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:          "arielsurco-cli",
	Short:        "Personal developer CLI",
	Version:      Version,
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(cmdproject.Cmd)
}
