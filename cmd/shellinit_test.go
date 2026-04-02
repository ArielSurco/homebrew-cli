package cmd_test

import (
	"strings"
	"testing"

	"github.com/arielsurco/go-cli/cmd"
	"github.com/arielsurco/go-cli/internal/module"
	"github.com/arielsurco/go-cli/internal/shell"
)

func TestRunShellInitWithShell_Bash(t *testing.T) {
	output := cmd.RunShellInitWithShell(shell.Bash, []string{"go-project"})
	if !strings.Contains(output, "gp()") {
		t.Errorf("expected bash output to contain 'gp()', got:\n%s", output)
	}
	if !strings.Contains(output, "complete -F _gp_completions gp") {
		t.Errorf("expected bash output to contain completion wiring for gp, got:\n%s", output)
	}
}

func TestRunShellInitWithShell_Zsh(t *testing.T) {
	output := cmd.RunShellInitWithShell(shell.Zsh, []string{"go-project"})
	if !strings.Contains(output, "gp()") {
		t.Errorf("expected zsh output to contain 'gp()', got:\n%s", output)
	}
	if !strings.Contains(output, "compdef _gp gp") {
		t.Errorf("expected zsh output to contain compdef wiring for gp, got:\n%s", output)
	}
}

func TestRunShellInitWithShell_EmptyActiveList(t *testing.T) {
	output := cmd.RunShellInitWithShell(shell.Bash, []string{})
	if output != "" {
		t.Errorf("expected empty output for empty active list, got %q", output)
	}
}

func TestRunShellInitWithShell_ContainsExpectedFunctionNames(t *testing.T) {
	output := cmd.RunShellInitWithShell(shell.Bash, []string{"go-project"})
	expectedFunctions := []string{"gp()", "gpd()", "gpa()", "gpr()"}
	for _, expectedFunction := range expectedFunctions {
		if !strings.Contains(output, expectedFunction) {
			t.Errorf("expected output to contain %q", expectedFunction)
		}
	}
}

func TestRunShellInitWithShell_InvalidShellFlag(t *testing.T) {
	// ParseShell is the gate for --shell flag validation — verify it rejects invalid values.
	_, err := shell.ParseShell("fish")
	if err == nil {
		t.Error("expected error for invalid shell value 'fish', got nil")
	}
}

func TestRunShellInitWithShell_BashOutputMatchesGenerate(t *testing.T) {
	activeNames := []string{"go-project"}
	output := cmd.RunShellInitWithShell(shell.Bash, activeNames)

	// Verify output matches Generate directly
	activeModules := module.ActiveModules(activeNames)
	expectedOutput := shell.Generate(activeModules, shell.Bash)
	if output != expectedOutput {
		t.Errorf("RunShellInitWithShell output does not match shell.Generate output\ngot:\n%s\nwant:\n%s", output, expectedOutput)
	}
}

func TestRunShellInitWithShell_BashDoesNotContainZshSyntax(t *testing.T) {
	output := cmd.RunShellInitWithShell(shell.Bash, []string{"go-project"})
	if strings.Contains(output, "compdef") {
		t.Error("bash output should not contain 'compdef' (zsh syntax)")
	}
}

func TestRunShellInitWithShell_ZshDoesNotContainBashSyntax(t *testing.T) {
	output := cmd.RunShellInitWithShell(shell.Zsh, []string{"go-project"})
	if strings.Contains(output, "complete -F") {
		t.Error("zsh output should not contain 'complete -F' (bash syntax)")
	}
}
