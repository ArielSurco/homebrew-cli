package shell_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/arielsurco/go-cli/internal/module"
	"github.com/arielsurco/go-cli/internal/shell"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestGenerate_Bash(t *testing.T) {
	output := shell.Generate(module.Registry, shell.Bash)
	goldenPath := filepath.Join("testdata", "golden_bash.txt")
	if *updateGolden {
		if err := os.WriteFile(goldenPath, []byte(output), 0o644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		return
	}
	expectedOutput, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}
	if string(expectedOutput) != output {
		t.Errorf("bash output does not match golden file\ngot:\n%s\nwant:\n%s", output, string(expectedOutput))
	}
}

func TestGenerate_Zsh(t *testing.T) {
	output := shell.Generate(module.Registry, shell.Zsh)
	goldenPath := filepath.Join("testdata", "golden_zsh.txt")
	if *updateGolden {
		if err := os.WriteFile(goldenPath, []byte(output), 0o644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		return
	}
	expectedOutput, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}
	if string(expectedOutput) != output {
		t.Errorf("zsh output does not match golden file\ngot:\n%s\nwant:\n%s", output, string(expectedOutput))
	}
}

func TestGenerate_EmptyModules_ReturnsEmptyString(t *testing.T) {
	output := shell.Generate([]module.Module{}, shell.Bash)
	if output != "" {
		t.Errorf("expected empty string for empty modules, got %q", output)
	}
}

func TestGenerate_SingleCommand_NoCompletions(t *testing.T) {
	singleModule := []module.Module{
		{
			Name: "test-module",
			Commands: []module.CommandDef{
				{
					Alias:          "mycmd",
					CobraCmd:       "some subcommand",
					NeedsEval:      false,
					HasCompletions: false,
					Description:    "a test command",
				},
			},
		},
	}

	output := shell.Generate(singleModule, shell.Bash)

	expectedOutput := "# arielsurco-cli shell wrappers — generated, do not edit manually\n\nmycmd() {\n  arielsurco-cli some subcommand \"$@\"\n}\n"
	if output != expectedOutput {
		t.Errorf("unexpected output for single command no-completions\ngot:\n%s\nwant:\n%s", output, expectedOutput)
	}
}
