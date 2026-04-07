package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

const shellInitContent = `eval "$(arielsurco-cli shell-init)"
`

const shellInitFileName = "shell-init.sh"

// CreateShellInitFile creates ~/.config/arielsurco-cli/shell-init.sh with the
// eval command. Returns the path to the created file. Uses XDG config path and
// writes atomically via tempfile + rename, matching config.go patterns.
func CreateShellInitFile() (string, error) {
	xdg.Reload()
	initDir := filepath.Join(xdg.ConfigHome, "arielsurco-cli")
	initPath := filepath.Join(initDir, shellInitFileName)

	if err := os.MkdirAll(initDir, 0o755); err != nil {
		return "", fmt.Errorf("creating config dir: %w", err)
	}

	tempFile, err := os.CreateTemp(initDir, "*.sh.tmp")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tempFilePath := tempFile.Name()

	if _, err := tempFile.Write([]byte(shellInitContent)); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFilePath)
		return "", fmt.Errorf("writing shell init: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempFilePath)
		return "", fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tempFilePath, initPath); err != nil {
		_ = os.Remove(tempFilePath)
		return "", fmt.Errorf("renaming shell init file: %w", err)
	}

	return initPath, nil
}

// RCFilePath returns the path to the shell's rc file for the given home directory.
func RCFilePath(targetShell Shell, homeDir string) (string, error) {
	switch targetShell {
	case Zsh:
		return filepath.Join(homeDir, ".zshrc"), nil
	case Bash:
		return filepath.Join(homeDir, ".bashrc"), nil
	default:
		return "", fmt.Errorf("unsupported shell for rc file")
	}
}

// shellInitBlock returns the block to inject into the rc file.
// initPath should use ~ instead of the literal home directory.
func shellInitBlock(tildeInitPath string) string {
	return fmt.Sprintf(`
# Load arielsurco-cli
if [[ -f "%s" ]]; then
  source "%s"
else
  echo "Failed while loading arielsurco-cli"
fi
`, tildeInitPath, tildeInitPath)
}

// replaceHomeWithTilde replaces the home directory prefix with ~ in the given path.
func replaceHomeWithTilde(path, homeDir string) string {
	if strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
	}
	return path
}

// InjectShellInit creates the shell init file and injects the source block into
// the user's rc file. Returns true if newly injected, false if already present.
func InjectShellInit(targetShell Shell) (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("getting home directory: %w", err)
	}
	return InjectShellInitWithHome(targetShell, homeDir)
}

// InjectShellInitWithHome is the testable core of InjectShellInit, accepting a
// custom home directory.
func InjectShellInitWithHome(targetShell Shell, homeDir string) (bool, error) {
	initPath, err := CreateShellInitFile()
	if err != nil {
		return false, fmt.Errorf("creating shell init file: %w", err)
	}

	rcPath, err := RCFilePath(targetShell, homeDir)
	if err != nil {
		return false, err
	}

	tildeInitPath := replaceHomeWithTilde(initPath, homeDir)
	marker := "arielsurco-cli/shell-init.sh"

	// Read existing rc file content (may not exist yet).
	existingContent, err := os.ReadFile(rcPath)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("reading rc file: %w", err)
	}

	// Check if already injected.
	if strings.Contains(string(existingContent), marker) {
		return false, nil
	}

	// Append the block.
	block := shellInitBlock(tildeInitPath)

	f, err := os.OpenFile(rcPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return false, fmt.Errorf("opening rc file: %w", err)
	}
	defer f.Close() //nolint:errcheck

	if _, err := f.WriteString(block); err != nil {
		return false, fmt.Errorf("writing to rc file: %w", err)
	}

	return true, nil
}
