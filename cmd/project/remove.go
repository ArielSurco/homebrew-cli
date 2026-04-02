package project

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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

func init() {
	removeCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
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

	skipConfirm, _ := cmd.Flags().GetBool("yes")
	return runRemoveWithTerminalState(projectName, shell.IsInteractiveSession(), skipConfirm, cfg)
}

// RunRemoveWithTerminalState is the testable core. Exported for testing.
func RunRemoveWithTerminalState(projectName string, isTerminal bool, skipConfirm bool, cfg *config.Config) error {
	return runRemoveWithTerminalState(projectName, isTerminal, skipConfirm, cfg)
}

func runRemoveWithTerminalState(projectName string, isTerminal bool, skipConfirm bool, cfg *config.Config) error {
	if len(cfg.Projects) == 0 {
		return fmt.Errorf("no projects configured\n\nAdd your first project:\n  gpa <name> <path>\n\nExample:\n  gpa my-app /Users/%s/projects/my-app", os.Getenv("USER"))
	}

	// Non-TTY without --yes: require explicit confirmation flag.
	if !isTerminal && !skipConfirm {
		return fmt.Errorf("removing project %q requires confirmation: use --yes to confirm in non-interactive mode", projectName)
	}

	// Non-TTY with --yes: requires exact name.
	if !isTerminal {
		if projectName == "" {
			return project.ErrNameRequired
		}
		return removeByName(projectName, cfg)
	}

	// TTY with direct name argument: smart navigation.
	if projectName != "" {
		svc := project.NewService(cfg)
		_, findErr := svc.FindByName(projectName)

		if findErr == nil {
			// Exact match found — confirm if needed, then remove.
			if !skipConfirm {
				confirmed, confirmErr := confirmRemove(projectName)
				if confirmErr != nil {
					return confirmErr
				}
				if !confirmed {
					return nil
				}
			}
			return removeByName(projectName, cfg)
		}

		if !errors.Is(findErr, project.ErrNotFound) {
			return findErr
		}

		// No exact match — check fuzzy matches to decide whether to open TUI.
		projectNames := make([]string, len(cfg.Projects))
		for index, existingProject := range cfg.Projects {
			projectNames[index] = existingProject.Name
		}
		if len(fuzzy.Find(projectName, projectNames)) == 0 {
			return fmt.Errorf("project %q not found", projectName)
		}
	}

	// TTY: launch TUI (empty filter or fuzzy-matched filter as seed).
	tty, err := shell.OpenTTY()
	if err != nil {
		return err
	}
	defer tty.Close() //nolint:errcheck

	tuiModel := projectlist.NewForDelete(cfg.Projects, projectName)
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

	switch selectionResult.Action {
	case projectlist.ActionDelete:
		// d→y sequence in TUI is the confirmation — skip prompt entirely.
		return removeByName(selectionResult.Project.Name, cfg)
	case projectlist.ActionNavigate:
		// User pressed Enter to select — still need CLI confirmation if not skipped.
		if !skipConfirm {
			confirmed, confirmErr := confirmRemove(selectionResult.Project.Name)
			if confirmErr != nil {
				return confirmErr
			}
			if !confirmed {
				return nil
			}
		}
		return removeByName(selectionResult.Project.Name, cfg)
	default:
		return nil
	}
}

// confirmRemoveFromReader reads a y/N confirmation from the provided reader.
// It returns true only when the user explicitly enters "y" or "Y".
func confirmRemoveFromReader(name string, reader io.Reader) (bool, error) {
	fmt.Fprintf(os.Stderr, "Remove project %q? [y/N]: ", name)
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return false, nil
	}
	return strings.ToLower(strings.TrimSpace(scanner.Text())) == "y", nil
}

// ConfirmRemoveFromReader is the exported version of confirmRemoveFromReader for testing.
func ConfirmRemoveFromReader(name string, reader io.Reader) (bool, error) {
	return confirmRemoveFromReader(name, reader)
}

// confirmRemove opens /dev/tty and prompts the user to confirm project removal.
func confirmRemove(name string) (bool, error) {
	tty, err := shell.OpenTTY()
	if err != nil {
		return false, err
	}
	defer tty.Close() //nolint:errcheck
	return confirmRemoveFromReader(name, tty)
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
	fmt.Fprintf(os.Stderr, "Project %q removed.\n", name)
	return nil
}
