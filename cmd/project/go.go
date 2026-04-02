package project

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/project"
	"github.com/ArielSurco/cli/internal/shell"
	"github.com/ArielSurco/cli/internal/tui/projectlist"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var goCmd = &cobra.Command{
	Use:               "go [project-name]",
	Short:             "Navigate to a project directory",
	Args:              cobra.MaximumNArgs(1),
	RunE:              runGo,
	ValidArgsFunction: completeProjectNames,
}

func runGo(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectName := ""
	if len(args) > 0 {
		projectName = args[0]
	}

	return runGoWithOutput(projectName, shell.IsInteractiveSession(), cfg, cmd.OutOrStdout())
}

// RunGoWithTerminalState is the testable core using stdout as output.
// Exported for testing.
func RunGoWithTerminalState(projectName string, isTerminal bool, cfg *config.Config) error {
	return runGoWithOutput(projectName, isTerminal, cfg, os.Stdout)
}

// RunGoWithOutput is the testable core with configurable output writer.
// Exported for testing.
func RunGoWithOutput(projectName string, isTerminal bool, cfg *config.Config, output io.Writer) error {
	return runGoWithOutput(projectName, isTerminal, cfg, output)
}

func runGoWithOutput(projectName string, isTerminal bool, cfg *config.Config, output io.Writer) error {
	if len(cfg.Projects) == 0 {
		return fmt.Errorf("no projects configured: use 'project add' to register one")
	}

	// Non-TTY: direct lookup path
	if !isTerminal {
		if projectName == "" {
			return project.ErrNameRequired
		}
		svc := project.NewService(cfg)
		foundProject, err := svc.FindByName(projectName)
		if err != nil {
			if errors.Is(err, project.ErrNotFound) {
				return fmt.Errorf("project %q not found", projectName)
			}
			return err
		}
		if _, err := fmt.Fprintln(output, foundProject.Path); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		return nil
	}

	// TTY: launch TUI with optional preFilter.
	// Open /dev/tty and configure lipgloss so styles render correctly even when
	// stdout is a pipe inside command substitution $(...).
	tty, err := shell.OpenTTY()
	if err != nil {
		return err
	}
	defer tty.Close()

	tuiModel := projectlist.New(cfg.Projects, projectName)
	finalProgram, err := tea.NewProgram(tuiModel,
		tea.WithAltScreen(),
		tea.WithOutput(tty),
		tea.WithInput(tty),
	).Run()
	if err != nil {
		return fmt.Errorf("running project selector: %w", err)
	}

	selectionResult := finalProgram.(projectlist.Model).Result()
	if selectionResult.Cancelled {
		return fmt.Errorf("cancelled")
	}

	if _, err := fmt.Fprintln(output, selectionResult.Project.Path); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

func completeProjectNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	projectNames := make([]string, len(cfg.Projects))
	for index, existingProject := range cfg.Projects {
		projectNames[index] = existingProject.Name
	}
	return projectNames, cobra.ShellCompDirectiveNoFileComp
}
