package project_test

import (
	"bytes"
	"strings"
	"testing"

	cmdproject "github.com/arielsurco/go-cli/cmd/project"
	"github.com/arielsurco/go-cli/internal/config"
)

// TestGoCommand_NonTTY_NoArgs verifies that in non-TTY mode without a project name,
// the command returns ErrNameRequired.
func TestGoCommand_NonTTY_NoArgs(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "go-cli", Path: "/projects/go-cli"},
		},
	}

	err := cmdproject.RunGoWithTerminalState("", false, cfg)
	if err == nil {
		t.Fatal("expected error for non-TTY with no args, got nil")
	}
	if !strings.Contains(err.Error(), "name required") && !strings.Contains(err.Error(), "project name") {
		t.Errorf("expected ErrNameRequired-related error, got: %v", err)
	}
}

// TestGoCommand_NonTTY_WithName verifies that in non-TTY mode with a project name,
// the command prints the project path to stdout.
func TestGoCommand_NonTTY_WithName(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "go-cli", Path: "/projects/go-cli"},
			{Name: "api", Path: "/projects/api"},
		},
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.RunGoWithOutput("go-cli", false, cfg, &outputBuffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(outputBuffer.String())
	if output != "/projects/go-cli" {
		t.Errorf("expected output /projects/go-cli, got %q", output)
	}
}

// TestGoCommand_NonTTY_NotFound verifies that in non-TTY mode with an unknown project name,
// the command returns ErrNotFound.
func TestGoCommand_NonTTY_NotFound(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "go-cli", Path: "/projects/go-cli"},
		},
	}

	err := cmdproject.RunGoWithTerminalState("unknown-project", false, cfg)
	if err == nil {
		t.Fatal("expected error for unknown project, got nil")
	}
}

// TestGoCommand_NonTTY_EmptyConfig verifies that in non-TTY mode with empty config,
// the command returns a descriptive error.
func TestGoCommand_NonTTY_EmptyConfig(t *testing.T) {
	cfg := &config.Config{Projects: []config.Project{}}

	err := cmdproject.RunGoWithTerminalState("some-project", false, cfg)
	if err == nil {
		t.Fatal("expected error for empty config, got nil")
	}
}

// TestGoCommand_TTY_EmptyConfig verifies that in TTY mode with empty config,
// the command returns a descriptive error before launching the TUI.
func TestGoCommand_TTY_EmptyConfig(t *testing.T) {
	cfg := &config.Config{Projects: []config.Project{}}

	err := cmdproject.RunGoWithTerminalState("", true, cfg)
	if err == nil {
		t.Fatal("expected error for TTY mode with empty config, got nil")
	}
}
