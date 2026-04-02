package cmd

import (
	"fmt"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/module"
	"github.com/ArielSurco/cli/internal/shell"
	"github.com/spf13/cobra"
)

// ShellInitCmd is the cobra command for shell-init. Exported for testing.
var ShellInitCmd = &cobra.Command{
	Use:   "shell-init",
	Short: "Print shell wrapper functions for active modules",
	RunE:  runShellInit,
}

func init() {
	ShellInitCmd.Flags().String("shell", "", "Shell type: bash or zsh (default: auto-detect from $SHELL)")
	rootCmd.AddCommand(ShellInitCmd)
}

func runShellInit(cobraCmd *cobra.Command, args []string) error {
	flagValue, err := cobraCmd.Flags().GetString("shell")
	if err != nil {
		return fmt.Errorf("reading --shell flag: %w", err)
	}

	var targetShell shell.Shell
	if flagValue != "" {
		targetShell, err = shell.ParseShell(flagValue)
		if err != nil {
			return err
		}
	} else {
		targetShell = shell.DetectShell()
	}

	activeModulesConfig, err := config.LoadActive()
	if err != nil {
		return fmt.Errorf("loading active modules: %w", err)
	}

	activeModules := module.ActiveModules(activeModulesConfig.Modules.Active)
	output := shell.Generate(activeModules, targetShell)
	if _, err = fmt.Fprint(cobraCmd.OutOrStdout(), output); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

// RunShellInitWithShell is the testable core of shell-init with an explicit shell type.
// Exported for testing.
func RunShellInitWithShell(targetShell shell.Shell, activeNames []string) string {
	activeModules := module.ActiveModules(activeNames)
	return shell.Generate(activeModules, targetShell)
}
