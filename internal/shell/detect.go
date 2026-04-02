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
