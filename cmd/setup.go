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
	defer tty.Close() //nolint:errcheck

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
	return applySetupResult(selectionResult.ActiveModules, selectionResult.Saved, "")
}

// RunSetupWithResult applies a pre-computed TUI result. Used for testing the save path.
// homeDir overrides os.UserHomeDir() when non-empty (for test isolation).
// Exported for testing.
func RunSetupWithResult(activeModuleNames []string, saved bool, homeDir string) error {
	return applySetupResult(activeModuleNames, saved, homeDir)
}

func applySetupResult(activeModuleNames []string, saved bool, homeDir string) error {
	if !saved {
		return nil
	}

	newActiveModules := &config.ActiveModules{
		Modules: config.ModulesSection{
			Active: activeModuleNames,
		},
	}
	if err := config.SaveActive(newActiveModules); err != nil {
		return fmt.Errorf("saving active modules: %w", err)
	}

	detectedShell := shell.DetectShell()

	var newlyInjected bool
	var err error
	if homeDir != "" {
		newlyInjected, err = shell.InjectShellInitWithHome(detectedShell, homeDir)
	} else {
		newlyInjected, err = shell.InjectShellInit(detectedShell)
	}
	if err != nil {
		fmt.Println("Setup saved. Run 'eval \"$(arielsurco-cli shell-init)\"' to apply.")
		return nil
	}

	if newlyInjected {
		rcName := "~/.bashrc"
		if detectedShell == shell.Zsh {
			rcName = "~/.zshrc"
		}
		fmt.Printf("Setup saved. Shell init added to your %s. Restart your shell to apply.\n", rcName)
	} else {
		fmt.Println("Setup saved. Restart your shell to apply.")
	}
	return nil
}
