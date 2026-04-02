package project

import (
	"errors"
	"fmt"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/project"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Unregister a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	svc := project.NewService(cfg)
	if err := svc.Remove(name); err != nil {
		if errors.Is(err, project.ErrNotFound) {
			return fmt.Errorf("project %q not found", name)
		}
		return err
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Project %q removed.\n", name)
	return nil
}
