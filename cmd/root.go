package cmd

import (
	"context"

	"charm.land/fang/v2"
	cmdproject "github.com/arielsurco/go-cli/cmd/project"
	"github.com/spf13/cobra"
)

// Version is injected at build time.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "arielsurco-cli",
	Short:   "Personal developer CLI",
	Version: Version,
}

// Execute runs the root command via fang.
func Execute() error {
	return fang.Execute(context.Background(), rootCmd)
}

func init() {
	rootCmd.AddCommand(cmdproject.Cmd)
}
