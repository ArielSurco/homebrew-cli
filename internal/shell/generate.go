package shell

import (
	"strings"

	"github.com/arielsurco/go-cli/internal/module"
)

// Generate emits shell wrapper functions for all commands in the given modules.
// Output is valid bash or zsh syntax depending on targetShell.
// Commands with HasCompletions=true also get completion wiring emitted.
// Returns an empty string when activeModules is empty.
func Generate(activeModules []module.Module, targetShell Shell) string {
	if len(activeModules) == 0 {
		return ""
	}

	var blocks []string

	for _, activeModule := range activeModules {
		for _, commandDef := range activeModule.Commands {
			block := buildCommandBlock(commandDef, targetShell)
			blocks = append(blocks, block)
		}
	}

	header := "# arielsurco-cli shell wrappers — generated, do not edit manually"
	allBlocks := append([]string{header}, blocks...)
	return strings.Join(allBlocks, "\n\n") + "\n"
}

// buildCommandBlock returns the shell block for a single CommandDef.
func buildCommandBlock(commandDef module.CommandDef, targetShell Shell) string {
	var sb strings.Builder

	// Function definition
	sb.WriteString(commandDef.Alias)
	sb.WriteString("() {\n")
	if commandDef.NeedsEval {
		sb.WriteString("  eval \"$(arielsurco-cli ")
		sb.WriteString(commandDef.CobraCmd)
		sb.WriteString(" \"$@\")\"\n")
	} else {
		sb.WriteString("  arielsurco-cli ")
		sb.WriteString(commandDef.CobraCmd)
		sb.WriteString(" \"$@\"\n")
	}
	sb.WriteString("}")

	// Completion wiring
	if commandDef.HasCompletions {
		sb.WriteString("\n")
		switch targetShell {
		case Bash:
			sb.WriteString(buildBashCompletion(commandDef))
		case Zsh:
			sb.WriteString(buildZshCompletion(commandDef))
		}
	}

	return sb.String()
}

// buildBashCompletion returns the bash completion wiring for a CommandDef.
func buildBashCompletion(commandDef module.CommandDef) string {
	var sb strings.Builder
	helperName := "_" + commandDef.Alias + "_completions"

	sb.WriteString(helperName)
	sb.WriteString("() {\n")
	sb.WriteString("  local currentWord=\"${COMP_WORDS[COMP_CWORD]}\"\n")
	sb.WriteString("  COMPREPLY=($(arielsurco-cli __complete ")
	sb.WriteString(commandDef.CobraCmd)
	sb.WriteString(" \"$currentWord\" 2>/dev/null | grep -v '^:'))\n")
	sb.WriteString("}\n")
	sb.WriteString("complete -F ")
	sb.WriteString(helperName)
	sb.WriteString(" ")
	sb.WriteString(commandDef.Alias)

	return sb.String()
}

// buildZshCompletion returns the zsh completion wiring for a CommandDef.
func buildZshCompletion(commandDef module.CommandDef) string {
	var sb strings.Builder
	helperName := "_" + commandDef.Alias

	sb.WriteString(helperName)
	sb.WriteString("() {\n")
	sb.WriteString("  local -a completionItems\n")
	sb.WriteString("  completionItems=($(arielsurco-cli __complete ")
	sb.WriteString(commandDef.CobraCmd)
	sb.WriteString(" \"${words[2]}\" 2>/dev/null | grep -v '^:'))\n")
	sb.WriteString("  _describe 'project' completionItems\n")
	sb.WriteString("}\n")
	sb.WriteString("compdef ")
	sb.WriteString(helperName)
	sb.WriteString(" ")
	sb.WriteString(commandDef.Alias)

	return sb.String()
}
