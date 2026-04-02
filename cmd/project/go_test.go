package project_test

import (
	"bytes"
	"strings"
	"testing"

	cmdproject "github.com/ArielSurco/cli/cmd/project"
	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/tui/projectlist"
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

// TestGoCommand_TTY_ExactMatch verifies that in TTY mode with an exact project name,
// the command prints the path directly without launching the TUI.
func TestGoCommand_TTY_ExactMatch(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "go-cli", Path: "/projects/go-cli"},
			{Name: "api", Path: "/projects/api"},
		},
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.RunGoWithOutput("go-cli", true, cfg, &outputBuffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(outputBuffer.String())
	if output != "/projects/go-cli" {
		t.Errorf("expected /projects/go-cli, got %q", output)
	}
}

// TestGoCommand_TTY_NoFuzzyMatch verifies that in TTY mode with a name that has no
// fuzzy matches, the command returns an error without launching the TUI.
func TestGoCommand_TTY_NoFuzzyMatch(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "go-cli", Path: "/projects/go-cli"},
			{Name: "api", Path: "/projects/api"},
		},
	}

	err := cmdproject.RunGoWithOutput("zzzznotfound", true, cfg, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for TTY with no fuzzy match, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestHandleGoResult_ActionNone verifies that ActionNone (cancelled) returns nil with no output.
func TestHandleGoResult_ActionNone(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "go-cli", Path: "/projects/go-cli"},
		},
	}

	result := projectlist.Result{
		Action:    projectlist.ActionNone,
		Cancelled: true,
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.HandleGoResult(result, cfg, &outputBuffer)
	if err != nil {
		t.Fatalf("expected nil error for ActionNone, got: %v", err)
	}
	if outputBuffer.Len() != 0 {
		t.Errorf("expected no output for ActionNone, got: %q", outputBuffer.String())
	}
}

// TestHandleGoResult_ActionNavigate verifies that ActionNavigate prints the project path.
func TestHandleGoResult_ActionNavigate(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "go-cli", Path: "/projects/go-cli"},
		},
	}

	result := projectlist.Result{
		Project: config.Project{Name: "go-cli", Path: "/projects/go-cli"},
		Action:  projectlist.ActionNavigate,
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.HandleGoResult(result, cfg, &outputBuffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(outputBuffer.String())
	if output != "/projects/go-cli" {
		t.Errorf("expected /projects/go-cli, got %q", output)
	}
}

// TestHandleGoResult_ActionDelete verifies that ActionDelete removes the project and saves config.
func TestHandleGoResult_ActionDelete(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "go-cli", Path: "/projects/go-cli"},
			{Name: "api", Path: "/projects/api"},
		},
	}

	result := projectlist.Result{
		Project: config.Project{Name: "go-cli", Path: "/projects/go-cli"},
		Action:  projectlist.ActionDelete,
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.HandleGoResult(result, cfg, &outputBuffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify project was removed from config.
	for _, proj := range cfg.Projects {
		if proj.Name == "go-cli" {
			t.Error("expected go-cli to be removed from config, but it still exists")
		}
	}
}

// TestHandleGoResult_ActionEditDev_NoEditor verifies that ActionEditDev returns an error
// when $EDITOR is not set.
func TestHandleGoResult_ActionEditDev_NoEditor(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("EDITOR", "")

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "go-cli", Path: "/projects/go-cli", DevScript: "make dev"},
		},
	}

	result := projectlist.Result{
		Project: config.Project{Name: "go-cli", Path: "/projects/go-cli", DevScript: "make dev"},
		Action:  projectlist.ActionEditDev,
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.HandleGoResult(result, cfg, &outputBuffer)
	if err == nil {
		t.Fatal("expected error when $EDITOR is not set, got nil")
	}
	if !strings.Contains(err.Error(), "$EDITOR is not set") {
		t.Errorf("expected '$EDITOR is not set' in error, got: %v", err)
	}
}
