package shell_test

import (
	"testing"

	"github.com/ArielSurco/cli/internal/shell"
)

func TestDetectShell_Zsh(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	if shell.DetectShell() != shell.Zsh {
		t.Error("expected Zsh when SHELL=/bin/zsh")
	}
}

func TestDetectShell_Bash(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	if shell.DetectShell() != shell.Bash {
		t.Error("expected Bash when SHELL=/bin/bash")
	}
}

func TestDetectShell_Unknown_DefaultsBash(t *testing.T) {
	t.Setenv("SHELL", "/bin/fish")
	if shell.DetectShell() != shell.Bash {
		t.Error("expected Bash as default for unknown shell")
	}
}

func TestDetectShell_Empty_DefaultsBash(t *testing.T) {
	t.Setenv("SHELL", "")
	if shell.DetectShell() != shell.Bash {
		t.Error("expected Bash as default when SHELL is empty")
	}
}

func TestParseShell_Valid(t *testing.T) {
	tests := []struct {
		name          string
		shellName     string
		expectedShell shell.Shell
	}{
		{name: "bash lowercase", shellName: "bash", expectedShell: shell.Bash},
		{name: "zsh lowercase", shellName: "zsh", expectedShell: shell.Zsh},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := shell.ParseShell(testCase.shellName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != testCase.expectedShell {
				t.Errorf("ParseShell(%q) = %v, want %v", testCase.shellName, result, testCase.expectedShell)
			}
		})
	}
}

func TestParseShell_Invalid(t *testing.T) {
	_, err := shell.ParseShell("fish")
	if err == nil {
		t.Error("expected error for unknown shell name, got nil")
	}
}

// Note: IsTerminal() depends on a real file descriptor and cannot be
// meaningfully unit-tested. It is covered by integration/acceptance tests only.
