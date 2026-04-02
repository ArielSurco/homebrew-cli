package project

import (
	"errors"
	"fmt"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/project"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <name> <path>",
	Short: "Register a new project",
	Args:  cobra.ExactArgs(2),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().String("dev-script", "", "Command to start the development server")
}

func runAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	path := args[1]

	devScript, err := cmd.Flags().GetString("dev-script")
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	svc := project.NewService(cfg)
	if err := svc.Add(name, path, devScript); err != nil {
		if errors.Is(err, project.ErrDuplicateName) {
			return fmt.Errorf("project %q already exists. Use 'gpr %s' first", name, name)
		}
		return err
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Project %q added.\n", name)
	return nil
}
