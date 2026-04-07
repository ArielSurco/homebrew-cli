package cmd_test

import (
	"testing"

	cmd "github.com/ArielSurco/cli/cmd"
	"github.com/ArielSurco/cli/internal/config"
)

// TestSetupCommand_NonTTY_ReturnsError verifies that in non-TTY mode
// the setup command returns an error requiring an interactive terminal.
func TestSetupCommand_NonTTY_ReturnsError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := cmd.RunSetupWithTerminalState(false)
	if err == nil {
		t.Fatal("expected error for non-TTY setup, got nil")
	}
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

// TestSetupCommand_Saved_SavesActiveModules verifies that when the TUI result
// has Saved=true, the active modules are written to config.
func TestSetupCommand_Saved_SavesActiveModules(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	activeModuleNames := []string{"go-project"}
	err := cmd.RunSetupWithResult(activeModuleNames, true, homeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the active modules were persisted
	savedConfig, loadErr := config.LoadActive()
	if loadErr != nil {
		t.Fatalf("failed to load active modules: %v", loadErr)
	}
	if len(savedConfig.Modules.Active) != 1 {
		t.Errorf("expected 1 active module saved, got %d", len(savedConfig.Modules.Active))
	}
	if savedConfig.Modules.Active[0] != "go-project" {
		t.Errorf("expected active module 'go-project', got %q", savedConfig.Modules.Active[0])
	}
}

// TestSetupCommand_Cancelled_ExitsSilently verifies that when the TUI result
// has Saved=false (cancelled), the command exits silently without error.
func TestSetupCommand_Cancelled_ExitsSilently(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := cmd.RunSetupWithResult([]string{}, false, "")
	if err != nil {
		t.Fatalf("expected nil when setup is cancelled, got: %v", err)
	}
}
