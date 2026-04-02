package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/project"
	"github.com/ArielSurco/cli/internal/shell"
	"github.com/ArielSurco/cli/internal/tui/projectscan"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [dir]",
	Short: "Scan a directory and register or unregister project subdirectories",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWatch,
}

func runWatch(cmd *cobra.Command, args []string) error {
	var rootPath string
	var err error

	if len(args) == 1 {
		rootPath, err = filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolving path: %w", err)
		}
	} else {
		rootPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	dirs, err := scanSubdirectories(rootPath)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		return fmt.Errorf("no subdirectories found in %s", rootPath)
	}

	registeredPaths := make(map[string]string, len(cfg.Projects))
	for _, proj := range cfg.Projects {
		registeredPaths[proj.Path] = proj.Name
	}

	if !shell.IsInteractiveSession() {
		return fmt.Errorf("watch requires an interactive terminal")
	}

	tty, err := shell.OpenTTY()
	if err != nil {
		return err
	}
	defer tty.Close() //nolint:errcheck

	tuiModel := projectscan.New(dirs, registeredPaths)
	finalProgram, err := tea.NewProgram(tuiModel,
		tea.WithInput(tty),
		tea.WithOutput(tty),
		tea.WithAltScreen(),
	).Run()
	if err != nil {
		return fmt.Errorf("running project scanner: %w", err)
	}

	result := finalProgram.(projectscan.Model).Result()
	if !result.Confirmed {
		return nil
	}

	return ApplyWatchResult(result, cfg)
}

// ScanSubdirectories returns direct subdirectories of rootPath, skipping hidden
// directories (names starting with "."), regular files, and symlinks.
// Exported for testing.
func ScanSubdirectories(rootPath string) ([]projectscan.DirEntry, error) {
	return scanSubdirectories(rootPath)
}

func scanSubdirectories(rootPath string) ([]projectscan.DirEntry, error) {
	rawEntries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", rootPath, err)
	}

	dirs := make([]projectscan.DirEntry, 0, len(rawEntries))
	for _, rawEntry := range rawEntries {
		if strings.HasPrefix(rawEntry.Name(), ".") {
			continue
		}
		if rawEntry.Type()&fs.ModeSymlink != 0 {
			continue
		}
		if !rawEntry.IsDir() {
			continue
		}
		dirs = append(dirs, projectscan.DirEntry{
			Name:    rawEntry.Name(),
			AbsPath: filepath.Join(rootPath, rawEntry.Name()),
		})
	}

	return dirs, nil
}

// ApplyWatchResult applies the TUI selection result to the config: adds new projects,
// removes deselected ones, and saves the config.
// Exported for testing.
func ApplyWatchResult(result projectscan.Result, cfg *config.Config) error {
	if !result.Confirmed {
		return nil
	}

	svc := project.NewService(cfg)

	for _, absPath := range result.ToAdd {
		dirName := filepath.Base(absPath)
		if addErr := svc.Add(dirName, absPath, ""); addErr != nil {
			if !errors.Is(addErr, project.ErrDuplicateName) {
				return addErr
			}
		}
	}

	for _, projectName := range result.ToRemove {
		if removeErr := svc.Remove(projectName); removeErr != nil {
			if !errors.Is(removeErr, project.ErrNotFound) {
				return removeErr
			}
		}
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Done. %d added, %d removed.\n", len(result.ToAdd), len(result.ToRemove))
	return nil
}
