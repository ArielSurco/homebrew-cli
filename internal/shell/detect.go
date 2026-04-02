package shell

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Shell represents a supported shell type.
type Shell int

const (
	// Bash is the GNU Bourne Again SHell.
	Bash Shell = iota
	// Zsh is the Z Shell.
	Zsh
)

// IsTerminal reports whether stdout is connected to a terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// IsInteractiveSession reports whether the user is running in an interactive
// terminal session by checking stdin. This returns true even inside command
// substitution $(...) where stdout is a pipe, making it suitable for deciding
// whether to launch a TUI.
func IsInteractiveSession() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// DetectShell reads $SHELL and returns the detected shell type.
// Defaults to Bash if unrecognized or unset.
func DetectShell() Shell {
	shellPath := os.Getenv("SHELL")
	shellName := strings.ToLower(strings.TrimSpace(shellPath))

	if strings.HasSuffix(shellName, "zsh") {
		return Zsh
	}
	return Bash
}

// ParseShell parses "bash" or "zsh" string into Shell type.
// Returns an error on unrecognized value.
func ParseShell(shellName string) (Shell, error) {
	switch strings.ToLower(strings.TrimSpace(shellName)) {
	case "bash":
		return Bash, nil
	case "zsh":
		return Zsh, nil
	default:
		return Bash, fmt.Errorf("unrecognized shell %q: supported values are bash, zsh", shellName)
	}
}
