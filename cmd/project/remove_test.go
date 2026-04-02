package project_test

import (
	"strings"
	"testing"

	cmdproject "github.com/ArielSurco/cli/cmd/project"
	"github.com/ArielSurco/cli/internal/config"
)

func TestRemoveCommand_NonTTY_ExactMatch(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
			{Name: "other", Path: "/path/to/other"},
		},
	}

	if err := cmdproject.RunRemoveWithTerminalState("myapp", false, true, cfg); err != nil {
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

	err := cmdproject.RunRemoveWithTerminalState("nonexistent", false, true, cfg)
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

	err := cmdproject.RunRemoveWithTerminalState("", false, false, cfg)
	if err == nil {
		t.Fatal("expected error for non-TTY with no args, got nil")
	}
}

func TestRemoveCommand_TTY_ExactMatch(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
			{Name: "other", Path: "/path/to/other"},
		},
	}

	if err := cmdproject.RunRemoveWithTerminalState("myapp", true, true, cfg); err != nil {
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

	err := cmdproject.RunRemoveWithTerminalState("zzzznotfound", true, false, cfg)
	if err == nil {
		t.Fatal("expected error for no fuzzy match, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestRemoveCommand_TTY_EmptyConfig(t *testing.T) {
	cfg := &config.Config{Projects: []config.Project{}}

	err := cmdproject.RunRemoveWithTerminalState("", true, false, cfg)
	if err == nil {
		t.Fatal("expected error for empty config, got nil")
	}
}

// --- New R2 scenarios ---

func TestRemoveCommand_TTY_DirectName_SkipConfirmTrue_Proceeds(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
			{Name: "other", Path: "/path/to/other"},
		},
	}

	err := cmdproject.RunRemoveWithTerminalState("myapp", true, true, cfg)
	if err != nil {
		t.Fatalf("expected deletion to proceed with --yes, got error: %v", err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project remaining, got %d", len(cfg.Projects))
	}
}

func TestRemoveCommand_TTY_DirectName_SkipConfirmFalse_UserConfirmsY_Proceeds(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
			{Name: "other", Path: "/path/to/other"},
		},
	}

	reader := strings.NewReader("y\n")
	confirmed, err := cmdproject.ConfirmRemoveFromReader("myapp", reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Fatal("expected confirmed=true when user enters 'y'")
	}

	// Now test the full flow with skipConfirm=true (simulating post-confirmation)
	err = cmdproject.RunRemoveWithTerminalState("myapp", true, true, cfg)
	if err != nil {
		t.Fatalf("expected deletion to proceed, got error: %v", err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project remaining, got %d", len(cfg.Projects))
	}
}

func TestRemoveCommand_TTY_DirectName_SkipConfirmFalse_UserEntersN_Aborts(t *testing.T) {
	reader := strings.NewReader("n\n")
	confirmed, err := cmdproject.ConfirmRemoveFromReader("myapp", reader)
	if err != nil {
		t.Fatalf("unexpected error from confirmRemoveFromReader: %v", err)
	}
	if confirmed {
		t.Fatal("expected confirmed=false when user enters 'n'")
	}
}

func TestRemoveCommand_NonTTY_SkipConfirmFalse_ReturnsUseYesError(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
		},
	}

	err := cmdproject.RunRemoveWithTerminalState("myapp", false, false, cfg)
	if err == nil {
		t.Fatal("expected error in non-TTY without --yes flag, got nil")
	}
	if !strings.Contains(err.Error(), "use --yes") {
		t.Errorf("expected error message to contain 'use --yes', got: %v", err)
	}
}

func TestRemoveCommand_NonTTY_SkipConfirmTrue_Proceeds(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
			{Name: "other", Path: "/path/to/other"},
		},
	}

	err := cmdproject.RunRemoveWithTerminalState("myapp", false, true, cfg)
	if err != nil {
		t.Fatalf("expected deletion to proceed with --yes in non-TTY, got error: %v", err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project remaining, got %d", len(cfg.Projects))
	}
}

func TestConfirmRemoveFromReader_YInput_ReturnsTrue(t *testing.T) {
	reader := strings.NewReader("y\n")
	confirmed, err := cmdproject.ConfirmRemoveFromReader("testproject", reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected true for 'y' input")
	}
}

func TestConfirmRemoveFromReader_YUppercase_ReturnsTrue(t *testing.T) {
	reader := strings.NewReader("Y\n")
	confirmed, err := cmdproject.ConfirmRemoveFromReader("testproject", reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected true for 'Y' input (case-insensitive)")
	}
}

func TestConfirmRemoveFromReader_NInput_ReturnsFalse(t *testing.T) {
	reader := strings.NewReader("n\n")
	confirmed, err := cmdproject.ConfirmRemoveFromReader("testproject", reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if confirmed {
		t.Error("expected false for 'n' input")
	}
}

func TestConfirmRemoveFromReader_EmptyInput_ReturnsFalse(t *testing.T) {
	reader := strings.NewReader("\n")
	confirmed, err := cmdproject.ConfirmRemoveFromReader("testproject", reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if confirmed {
		t.Error("expected false for empty input (default no)")
	}
}

func TestRemoveCommand_NonTTY_NoName_SkipConfirmTrue_ReturnsErrNameRequired(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/path/to/app"},
		},
	}

	err := cmdproject.RunRemoveWithTerminalState("", false, true, cfg)
	if err == nil {
		t.Fatal("expected ErrNameRequired when no name provided in non-TTY mode, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to reference missing name, got: %v", err)
	}
}

func TestConfirmRemoveFromReader_EOFInput_ReturnsFalse(t *testing.T) {
	reader := strings.NewReader("")
	confirmed, err := cmdproject.ConfirmRemoveFromReader("testproject", reader)
	if err != nil {
		t.Fatalf("unexpected error on EOF: %v", err)
	}
	if confirmed {
		t.Error("expected false for EOF (no input)")
	}
}
