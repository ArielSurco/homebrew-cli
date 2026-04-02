package project_test

import (
	"strings"
	"testing"

	cmdproject "github.com/ArielSurco/cli/cmd/project"
	"github.com/ArielSurco/cli/internal/config"
)

func TestRemoveCommand_NonTTY_ExactMatch(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
			{Name: "other", Path: "/path/to/other"},
		},
	}

	if err := cmdproject.RunRemoveWithTerminalState("myapp", false, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project after remove, got %d", len(cfg.Projects))
	}
	if cfg.Projects[0].Name != "other" {
		t.Errorf("expected 'other' to remain, got %q", cfg.Projects[0].Name)
	}
}

func TestRemoveCommand_NonTTY_NotFound(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
		},
	}

	err := cmdproject.RunRemoveWithTerminalState("nonexistent", false, cfg)
	if err == nil {
		t.Fatal("expected error for nonexistent project, got nil")
	}
}

func TestRemoveCommand_NonTTY_NoArgs(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
		},
	}

	err := cmdproject.RunRemoveWithTerminalState("", false, cfg)
	if err == nil {
		t.Fatal("expected error for non-TTY with no args, got nil")
	}
}

func TestRemoveCommand_TTY_ExactMatch(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
			{Name: "other", Path: "/path/to/other"},
		},
	}

	if err := cmdproject.RunRemoveWithTerminalState("myapp", true, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project after remove, got %d", len(cfg.Projects))
	}
}

func TestRemoveCommand_TTY_NoFuzzyMatch(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
		},
	}

	err := cmdproject.RunRemoveWithTerminalState("zzzznotfound", true, cfg)
	if err == nil {
		t.Fatal("expected error for no fuzzy match, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestRemoveCommand_TTY_EmptyConfig(t *testing.T) {
	cfg := &config.Config{Projects: []config.Project{}}

	err := cmdproject.RunRemoveWithTerminalState("", true, cfg)
	if err == nil {
		t.Fatal("expected error for empty config, got nil")
	}
}
