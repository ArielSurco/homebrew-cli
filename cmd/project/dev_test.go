package project_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	cmdproject "github.com/ArielSurco/cli/cmd/project"
	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/project"
	"github.com/ArielSurco/cli/internal/tui/projectlist"
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

// TestHandleDevResult_ActionNone verifies that ActionNone (cancelled) returns nil with no output.
func TestHandleDevResult_ActionNone(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/home/user/myapp", DevScript: "npm run dev"},
		},
	}

	result := projectlist.Result{
		Action:    projectlist.ActionNone,
		Cancelled: true,
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.HandleDevResult(result, cfg, &outputBuffer)
	if err != nil {
		t.Fatalf("expected nil error for ActionNone, got: %v", err)
	}
	if outputBuffer.Len() != 0 {
		t.Errorf("expected no output for ActionNone, got: %q", outputBuffer.String())
	}
}

// TestHandleDevResult_ActionDelete verifies that ActionDelete removes the project and saves config.
func TestHandleDevResult_ActionDelete(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/home/user/myapp", DevScript: "npm run dev"},
			{Name: "api", Path: "/home/user/api", DevScript: "go run ."},
		},
	}

	result := projectlist.Result{
		Project: config.Project{Name: "myapp", Path: "/home/user/myapp", DevScript: "npm run dev"},
		Action:  projectlist.ActionDelete,
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.HandleDevResult(result, cfg, &outputBuffer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify project was removed from config.
	for _, proj := range cfg.Projects {
		if proj.Name == "myapp" {
			t.Error("expected myapp to be removed from config, but it still exists")
		}
	}
}

// TestHandleDevResult_ActionEditDev_NoEditor verifies that ActionEditDev returns an error
// when $EDITOR is not set.
func TestHandleDevResult_ActionEditDev_NoEditor(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("EDITOR", "")

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myapp", Path: "/home/user/myapp", DevScript: "npm run dev"},
		},
	}

	result := projectlist.Result{
		Project: config.Project{Name: "myapp", Path: "/home/user/myapp", DevScript: "npm run dev"},
		Action:  projectlist.ActionEditDev,
	}

	var outputBuffer bytes.Buffer
	err := cmdproject.HandleDevResult(result, cfg, &outputBuffer)
	if err == nil {
		t.Fatal("expected error when $EDITOR is not set, got nil")
	}
	if !strings.Contains(err.Error(), "$EDITOR is not set") {
		t.Errorf("expected '$EDITOR is not set' in error, got: %v", err)
	}
}
