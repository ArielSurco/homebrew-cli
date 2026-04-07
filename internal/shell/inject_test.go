package shell_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ArielSurco/cli/internal/shell"
)

func TestCreateShellInitFile(t *testing.T) {
	homeDir := t.TempDir()

	initPath, err := shell.CreateShellInitFile(homeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := filepath.Join(homeDir, ".config", "arielsurco-cli", "shell-init.sh")
	if initPath != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, initPath)
	}

	content, err := os.ReadFile(initPath)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	expectedContent := `eval "$(arielsurco-cli shell-init)"` + "\n"
	if string(content) != expectedContent {
		t.Errorf("unexpected content:\ngot:  %q\nwant: %q", string(content), expectedContent)
	}
}

func TestRCFilePath_Bash(t *testing.T) {
	homeDir := "/home/testuser"
	rcPath, err := shell.RCFilePath(shell.Bash, homeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(homeDir, ".bashrc")
	if rcPath != expected {
		t.Errorf("expected %q, got %q", expected, rcPath)
	}
}

func TestRCFilePath_Zsh(t *testing.T) {
	homeDir := "/home/testuser"
	rcPath, err := shell.RCFilePath(shell.Zsh, homeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(homeDir, ".zshrc")
	if rcPath != expected {
		t.Errorf("expected %q, got %q", expected, rcPath)
	}
}

func TestInjectShellInit_NewInjection(t *testing.T) {
	homeDir := t.TempDir()

	// Create an empty rc file.
	rcPath := filepath.Join(homeDir, ".zshrc")
	if err := os.WriteFile(rcPath, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create rc file: %v", err)
	}

	newlyInjected, err := shell.InjectShellInitWithHome(shell.Zsh, homeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newlyInjected {
		t.Error("expected newly injected to be true")
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("failed to read rc file: %v", err)
	}

	if !strings.Contains(string(content), "arielsurco-cli/shell-init.sh") {
		t.Error("expected rc file to contain shell-init.sh source block")
	}
	if !strings.Contains(string(content), "# Load arielsurco-cli") {
		t.Error("expected rc file to contain load comment")
	}
}

func TestInjectShellInit_AlreadyInjected(t *testing.T) {
	homeDir := t.TempDir()

	// Create rc file with the marker already present.
	rcPath := filepath.Join(homeDir, ".zshrc")
	existingContent := "# existing stuff\nsource ~/.config/arielsurco-cli/shell-init.sh\n"
	if err := os.WriteFile(rcPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("failed to create rc file: %v", err)
	}

	newlyInjected, err := shell.InjectShellInitWithHome(shell.Zsh, homeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newlyInjected {
		t.Error("expected newly injected to be false when already present")
	}

	// Verify no duplicate was added.
	content, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("failed to read rc file: %v", err)
	}
	if string(content) != existingContent {
		t.Errorf("rc file was modified when it shouldn't have been:\ngot:  %q\nwant: %q", string(content), existingContent)
	}
}

func TestInjectShellInit_RCFileDoesNotExist(t *testing.T) {
	homeDir := t.TempDir()

	// Do NOT create the rc file — it should be created by InjectShellInit.
	newlyInjected, err := shell.InjectShellInitWithHome(shell.Zsh, homeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newlyInjected {
		t.Error("expected newly injected to be true")
	}

	rcPath := filepath.Join(homeDir, ".zshrc")
	content, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("rc file was not created: %v", err)
	}
	if !strings.Contains(string(content), "arielsurco-cli/shell-init.sh") {
		t.Error("expected created rc file to contain shell-init.sh source block")
	}
}

func TestInjectShellInit_Idempotent(t *testing.T) {
	homeDir := t.TempDir()

	rcPath := filepath.Join(homeDir, ".zshrc")
	if err := os.WriteFile(rcPath, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create rc file: %v", err)
	}

	// First call: should inject.
	firstResult, err := shell.InjectShellInitWithHome(shell.Zsh, homeDir)
	if err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	if !firstResult {
		t.Error("first call: expected newly injected to be true")
	}

	contentAfterFirst, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("failed to read rc file: %v", err)
	}

	// Second call: should not inject again.
	secondResult, err := shell.InjectShellInitWithHome(shell.Zsh, homeDir)
	if err != nil {
		t.Fatalf("second call: unexpected error: %v", err)
	}
	if secondResult {
		t.Error("second call: expected newly injected to be false")
	}

	contentAfterSecond, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("failed to read rc file: %v", err)
	}

	if string(contentAfterFirst) != string(contentAfterSecond) {
		t.Errorf("content changed after second call:\nfirst:  %q\nsecond: %q", string(contentAfterFirst), string(contentAfterSecond))
	}
}

func TestInjectShellInit_PreservesExistingContent(t *testing.T) {
	homeDir := t.TempDir()

	rcPath := filepath.Join(homeDir, ".bashrc")
	existingContent := "export PATH=/usr/local/bin:$PATH\nalias ll='ls -la'\n"
	if err := os.WriteFile(rcPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("failed to create rc file: %v", err)
	}

	newlyInjected, err := shell.InjectShellInitWithHome(shell.Bash, homeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newlyInjected {
		t.Error("expected newly injected to be true")
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("failed to read rc file: %v", err)
	}

	contentStr := string(content)

	// Original content must still be there.
	if !strings.HasPrefix(contentStr, existingContent) {
		t.Error("existing content was not preserved at the beginning of the file")
	}

	// New block must be appended.
	if !strings.Contains(contentStr, "arielsurco-cli/shell-init.sh") {
		t.Error("expected rc file to contain shell-init.sh source block")
	}
}

func TestInjectShellInit_ShellInitFileCreatedInFixedPath(t *testing.T) {
	homeDir := t.TempDir()

	newlyInjected, err := shell.InjectShellInitWithHome(shell.Bash, homeDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !newlyInjected {
		t.Error("expected newly injected to be true")
	}

	// Verify shell-init.sh is in ~/.config/arielsurco-cli/, not XDG.
	expectedInitPath := filepath.Join(homeDir, ".config", "arielsurco-cli", "shell-init.sh")
	if _, err := os.Stat(expectedInitPath); os.IsNotExist(err) {
		t.Errorf("expected shell-init.sh at %q, but it does not exist", expectedInitPath)
	}

	// Verify the rc file references $HOME/.config path.
	rcPath := filepath.Join(homeDir, ".bashrc")
	content, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("failed to read rc file: %v", err)
	}
	if !strings.Contains(string(content), "$HOME/.config/arielsurco-cli/shell-init.sh") {
		t.Errorf("expected rc file to reference $HOME/.config path, got:\n%s", string(content))
	}
}
