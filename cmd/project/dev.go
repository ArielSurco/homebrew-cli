package project

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/arielsurco/go-cli/internal/config"
	"github.com/arielsurco/go-cli/internal/project"
	"github.com/arielsurco/go-cli/internal/shell"
	"github.com/arielsurco/go-cli/internal/tui/projectlist"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:               "dev [project-name]",
	Short:             "Print the dev command for a project (for shell eval)",
	Args:              cobra.MaximumNArgs(1),
	RunE:              runDev,
	ValidArgsFunction: completeProjectNames,
}

func runDev(cobraCmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectName := ""
	if len(args) > 0 {
		projectName = args[0]
	}

	return runDevWithOutput(projectName, shell.IsTerminal(), cfg, cobraCmd.OutOrStdout())
}

// RunDevWithTerminalState is the testable core using stdout as output.
// Exported for testing.
func RunDevWithTerminalState(projectName string, isTerminal bool, cfg *config.Config) error {
	return runDevWithOutput(projectName, isTerminal, cfg, os.Stdout)
}

// RunDevWithOutput is the testable core with configurable output writer.
// Exported for testing.
func RunDevWithOutput(projectName string, isTerminal bool, cfg *config.Config, output io.Writer) error {
	return runDevWithOutput(projectName, isTerminal, cfg, output)
}

func runDevWithOutput(projectName string, isTerminal bool, cfg *config.Config, output io.Writer) error {
	if len(cfg.Projects) == 0 {
		return fmt.Errorf("no projects configured: use 'project add' to register one")
	}

	// Non-TTY: direct lookup path
	if !isTerminal {
		if projectName == "" {
			return project.ErrNameRequired
		}
		svc := project.NewService(cfg)
		shellCommand, err := svc.DevCommand(projectName)
		if err != nil {
			if errors.Is(err, project.ErrNotFound) {
				return fmt.Errorf("project %q not found", projectName)
			}
			if errors.Is(err, project.ErrNoDevScript) {
				return fmt.Errorf("%w: add a dev_script to your config or .arielsurco-cli.toml", project.ErrNoDevScript)
			}
			return err
		}
		if _, err := fmt.Fprintln(output, shellCommand); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		return nil
	}

	// TTY: launch TUI with optional preFilter
	tuiModel := projectlist.New(cfg.Projects, projectName)
	finalProgram, err := tea.NewProgram(tuiModel).Run()
	if err != nil {
		return fmt.Errorf("running project selector: %w", err)
	}

	selectionResult := finalProgram.(projectlist.Model).Result()
	if selectionResult.Cancelled {
		return fmt.Errorf("cancelled")
	}

	svc := project.NewService(cfg)
	shellCommand, err := svc.DevCommand(selectionResult.Project.Name)
	if err != nil {
		if errors.Is(err, project.ErrNoDevScript) {
			return fmt.Errorf("%w: add a dev_script to your config or .arielsurco-cli.toml", project.ErrNoDevScript)
		}
		return err
	}

	if _, err := fmt.Fprintln(output, shellCommand); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}
