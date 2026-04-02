package cmd

import (
	"fmt"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/module"
	"github.com/ArielSurco/cli/internal/shell"
	"github.com/ArielSurco/cli/internal/tui/setup"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Select which command modules to activate",
	RunE:  runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cobraCmd *cobra.Command, args []string) error {
	return RunSetupWithTerminalState(shell.IsTerminal())
}

// RunSetupWithTerminalState is the testable core for the non-TTY path.
// Exported for testing.
func RunSetupWithTerminalState(isTerminal bool) error {
	if !isTerminal {
		return fmt.Errorf("setup requires an interactive terminal")
	}

	currentActive, err := config.LoadActive()
	if err != nil {
		return fmt.Errorf("loading active modules: %w", err)
	}

	tty, err := shell.OpenTTY()
	if err != nil {
		return err
	}
	defer tty.Close()

	setupModel := setup.New(module.Registry, currentActive.Modules.Active)
	finalProgram, err := tea.NewProgram(setupModel,
		tea.WithAltScreen(),
		tea.WithOutput(tty),
		tea.WithInput(tty),
	).Run()
	if err != nil {
		return fmt.Errorf("running setup TUI: %w", err)
	}

	selectionResult := finalProgram.(setup.Model).Result()
	return applySetupResult(selectionResult.ActiveModules, selectionResult.Saved)
}

// RunSetupWithResult applies a pre-computed TUI result. Used for testing the save path.
// Exported for testing.
func RunSetupWithResult(activeModuleNames []string, saved bool) error {
	return applySetupResult(activeModuleNames, saved)
}

func applySetupResult(activeModuleNames []string, saved bool) error {
	if !saved {
		return fmt.Errorf("setup cancelled")
	}

	newActiveModules := &config.ActiveModules{
		Modules: config.ModulesSection{
			Active: activeModuleNames,
		},
	}
	if err := config.SaveActive(newActiveModules); err != nil {
		return fmt.Errorf("saving active modules: %w", err)
	}

	fmt.Println("Setup saved. Run 'eval \"$(arielsurco-cli shell-init)\"' to apply.")
	return nil
}
