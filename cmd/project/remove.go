package project

import (
	"errors"
	"fmt"
	"os"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/project"
	"github.com/ArielSurco/cli/internal/shell"
	"github.com/ArielSurco/cli/internal/tui/projectlist"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:               "remove [project-name]",
	Short:             "Unregister a project",
	Args:              cobra.MaximumNArgs(1),
	RunE:              runRemove,
	ValidArgsFunction: completeProjectNames,
}

func runRemove(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectName := ""
	if len(args) > 0 {
		projectName = args[0]
	}

	return runRemoveWithTerminalState(projectName, shell.IsInteractiveSession(), cfg)
}

// RunRemoveWithTerminalState is the testable core. Exported for testing.
func RunRemoveWithTerminalState(projectName string, isTerminal bool, cfg *config.Config) error {
	return runRemoveWithTerminalState(projectName, isTerminal, cfg)
}

func runRemoveWithTerminalState(projectName string, isTerminal bool, cfg *config.Config) error {
	if len(cfg.Projects) == 0 {
		return fmt.Errorf("no projects configured\n\nAdd your first project:\n  gpa <name> <path>\n\nExample:\n  gpa my-app /Users/%s/projects/my-app", os.Getenv("USER"))
	}

	// Non-TTY: requires exact name.
	if !isTerminal {
		if projectName == "" {
			return project.ErrNameRequired
		}
		return removeByName(projectName, cfg)
	}

	// TTY with argument: smart navigation.
	if projectName != "" {
		// 1. Exact match → remove directly without TUI.
		if err := removeByName(projectName, cfg); err == nil {
			return nil
		} else if !errors.Is(err, project.ErrNotFound) {
			return err
		}

		// 2. Check fuzzy matches.
		projectNames := make([]string, len(cfg.Projects))
		for index, existingProject := range cfg.Projects {
			projectNames[index] = existingProject.Name
		}
		if len(fuzzy.Find(projectName, projectNames)) == 0 {
			// 3. No matches → error.
			return fmt.Errorf("project %q not found", projectName)
		}
	}

	// TTY: launch TUI (empty filter or fuzzy-matched filter as seed).
	tty, err := shell.OpenTTY()
	if err != nil {
		return err
	}
	defer tty.Close() //nolint:errcheck

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
		return nil
	}

	return removeByName(selectionResult.Project.Name, cfg)
}

// removeByName removes the project from config and saves atomically.
func removeByName(name string, cfg *config.Config) error {
	svc := project.NewService(cfg)
	if err := svc.Remove(name); err != nil {
		if errors.Is(err, project.ErrNotFound) {
			return fmt.Errorf("%w: %q", project.ErrNotFound, name)
		}
		return err
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Printf("Project %q removed.\n", name)
	return nil
}
