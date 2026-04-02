package project_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	cmdproject "github.com/ArielSurco/cli/cmd/project"
	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/project"
)

// TestDevCommand_NonTTY_WithArg_Found verifies that in non-TTY mode with a known
// project name, the command prints cd '<path>' && <dev_script> to stdout.
func TestDevCommand_NonTTY_WithArg_Found(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/home/user/myapp", DevScript: "npm run dev"},
		},
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.RunDevWithOutput("myapp", false, cfg, &outputBuffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(outputBuffer.String())
	expectedCommand := "cd '/home/user/myapp' && npm run dev"
	if output != expectedCommand {
		t.Errorf("got %q, want %q", output, expectedCommand)
	}
}

// TestDevCommand_NonTTY_WithArg_NotFound verifies that in non-TTY mode with an
// unknown project name, the command returns an error.
func TestDevCommand_NonTTY_WithArg_NotFound(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/home/user/myapp", DevScript: "npm run dev"},
		},
	}

	err := cmdproject.RunDevWithOutput("unknown-project", false, cfg, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for unknown project, got nil")
	}
}

// TestDevCommand_NonTTY_NoArg verifies that in non-TTY mode without a project name,
// the command returns ErrNameRequired.
func TestDevCommand_NonTTY_NoArg(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/home/user/myapp", DevScript: "npm run dev"},
		},
	}

	err := cmdproject.RunDevWithOutput("", false, cfg, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for non-TTY with no args, got nil")
	}
	if !errors.Is(err, project.ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got: %v", err)
	}
}

// TestDevCommand_NonTTY_NoDevScript verifies that in non-TTY mode when a project
// has no dev_script configured, the command returns ErrNoDevScript.
func TestDevCommand_NonTTY_NoDevScript(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/home/user/myapp", DevScript: ""},
		},
	}

	err := cmdproject.RunDevWithOutput("myapp", false, cfg, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for project without dev_script, got nil")
	}
	if !errors.Is(err, project.ErrNoDevScript) {
		t.Errorf("expected ErrNoDevScript, got: %v", err)
	}
}

// TestDevCommand_NonTTY_PathWithSpaces verifies that paths containing spaces are
// correctly wrapped in single quotes.
func TestDevCommand_NonTTY_PathWithSpaces(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/home/my user/my app", DevScript: "make dev"},
		},
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.RunDevWithOutput("myapp", false, cfg, &outputBuffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(outputBuffer.String())
	expectedCommand := "cd '/home/my user/my app' && make dev"
	if output != expectedCommand {
		t.Errorf("got %q, want %q", output, expectedCommand)
	}
}

// TestDevCommand_NonTTY_PathWithSingleQuote verifies that single quotes in the path
// are escaped using the POSIX '\'' technique.
func TestDevCommand_NonTTY_PathWithSingleQuote(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/home/it's/here", DevScript: "yarn dev"},
		},
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.RunDevWithOutput("myapp", false, cfg, &outputBuffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(outputBuffer.String())
	expectedCommand := "cd '/home/it'\\''s/here' && yarn dev"
	if output != expectedCommand {
		t.Errorf("got %q, want %q", output, expectedCommand)
	}
}

// TestDevCommand_TTY_EmptyConfig verifies that in TTY mode with empty config,
// the command returns a descriptive error before launching the TUI.
func TestDevCommand_TTY_EmptyConfig(t *testing.T) {
	cfg := &config.Config{Projects: []config.Project{}}

	err := cmdproject.RunDevWithOutput("", true, cfg, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for TTY mode with empty config, got nil")
	}
}
