package project

import "github.com/spf13/cobra"

// devCmd is a stub for Phase 2/3.
var devCmd = &cobra.Command{
	Use:   "dev <name>",
	Short: "Start the dev server for a project (Phase 2/3)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
