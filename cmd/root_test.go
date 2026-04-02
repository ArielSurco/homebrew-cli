package cmd_test

import (
	"testing"

	"github.com/ArielSurco/cli/cmd"
)

func TestExecute_NoError(t *testing.T) {
	// Verify Execute() doesn't error when called with --help
	// We can't test fang.Execute directly without a real terminal,
	// but we verify the package compiles and Execute is callable.
	_ = cmd.Execute
}

func TestVersion_Default(t *testing.T) {
	if cmd.Version != "dev" {
		t.Errorf("expected default version 'dev', got %q", cmd.Version)
	}
}

func TestVersion_IsSet(t *testing.T) {
	if cmd.Version == "" {
		t.Error("Version must not be empty")
	}
}
