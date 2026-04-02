package project

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/project"
	"github.com/ArielSurco/cli/internal/shell"
	"github.com/ArielSurco/cli/internal/tui/projectlist"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
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
		return fmt.Errorf("no projects configured\n\nAdd your first project:\n  gpa <name> <path>\n\nExample:\n  gpa my-app /Users/%s/projects/my-app", os.Getenv("USER"))
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

	// TTY with argument: smart navigation.
	if projectName != "" {
		svc := project.NewService(cfg)

		// 1. Exact match → cd directly, no TUI.
		foundProject, err := svc.FindByName(projectName)
		if err == nil {
			if _, err := fmt.Fprintln(output, foundProject.Path); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}
			return nil
		}

		// 2. Fuzzy matches exist → open TUI with filter pre-applied.
		projectNames := make([]string, len(cfg.Projects))
		for index, existingProject := range cfg.Projects {
			projectNames[index] = existingProject.Name
		}
		if len(fuzzy.Find(projectName, projectNames)) == 0 {
			// 3. No matches at all → error.
			return fmt.Errorf("project %q not found", projectName)
		}
	}

	// TTY: launch TUI (empty filter or fuzzy-matched filter as seed).
	// Open /dev/tty and configure lipgloss so styles render correctly even when
	// stdout is a pipe inside command substitution $(...).
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
	return HandleGoResult(selectionResult, cfg, output)
}

// HandleGoResult processes the TUI result for the go command.
// Exported so tests can inject a pre-built result without launching the TUI.
func HandleGoResult(selectionResult projectlist.Result, cfg *config.Config, output io.Writer) error {
	switch selectionResult.Action {
	case projectlist.ActionNone:
		return nil
	case projectlist.ActionNavigate:
		if _, err := fmt.Fprintln(output, selectionResult.Project.Path); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		return nil
	case projectlist.ActionDelete:
		return removeByName(selectionResult.Project.Name, cfg)
	case projectlist.ActionEditDev:
		return runEditDevScript(selectionResult.Project, cfg, nil)
	default:
		return nil
	}
}

// runEditDevScript opens the project's dev script in $EDITOR for inline editing.
// tty is the terminal file used for editor I/O; if nil, os.Stdin/Stdout/Stderr are used.
func runEditDevScript(proj config.Project, cfg *config.Config, tty *os.File) error {
	editorEnv := os.Getenv("EDITOR")
	if editorEnv == "" {
		return fmt.Errorf("$EDITOR is not set; set it to your preferred editor (e.g. export EDITOR=vim)")
	}

	tempFile, err := os.CreateTemp("", "dev-script-*.sh")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) //nolint:errcheck

	if _, err := tempFile.WriteString(proj.DevScript); err != nil {
		tempFile.Close() //nolint:errcheck
		return fmt.Errorf("writing dev script to temp file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	editorFields := strings.Fields(editorEnv)
	editorArgs := make([]string, 0, len(editorFields)-1+1)
	editorArgs = append(editorArgs, editorFields[1:]...)
	editorArgs = append(editorArgs, tempFile.Name())
	editorCmd := exec.Command(editorFields[0], editorArgs...) //nolint:gosec

	if tty != nil {
		editorCmd.Stdin = tty
		editorCmd.Stdout = tty
		editorCmd.Stderr = tty
	} else {
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
	}

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("running editor: %w", err)
	}

	contents, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return fmt.Errorf("reading edited dev script: %w", err)
	}

	trimmedScript := strings.TrimSpace(string(contents))

	svc := project.NewService(cfg)
	if err := svc.UpdateDevScript(proj.Name, trimmedScript); err != nil {
		return fmt.Errorf("updating dev script: %w", err)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Dev script for %q updated.\n", proj.Name)
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
