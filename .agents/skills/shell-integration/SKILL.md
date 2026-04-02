---
name: shell-integration
description: >
  Shell wrapper generation, TTY detection, eval rules, and OpenTTY pattern for arielsurco-cli.
  Trigger: When modifying shell-init, wrappers, CommandDef, TTY detection, or adding commands
  that interact with the parent shell.
metadata:
  author: arielsurco
  version: "1.0"
---

## When to Use

- Modifying `internal/shell/detect.go` or `generate.go`
- Adding a new command that needs a shell alias
- Changing how wrappers handle cd, eval, or completions
- Debugging why a TUI has no styles or doesn't open inside `$()`

---

## Critical Patterns

### Pattern 1: TTY Detection — MUST use stdin, not stdout

```go
// ✅ CORRECT — works inside $() where stdout is a pipe
shell.IsInteractiveSession()  // term.IsTerminal(int(os.Stdin.Fd()))

// ❌ WRONG for TUI decisions — inside $() stdout is a pipe → always false
shell.IsTerminal()            // term.IsTerminal(int(os.Stdout.Fd()))
```

**Why**: When the shell wrapper runs `gp`, it uses `$()` (command substitution). Inside `$()`, stdout is a pipe but stdin is still the user's terminal. `IsInteractiveSession` correctly returns true.

### Pattern 2: OpenTTY — Required for TUI inside `$()`

```go
tty, err := shell.OpenTTY()  // opens /dev/tty + sets lipgloss color profile
if err != nil {
    return err
}
defer tty.Close() //nolint:errcheck

finalProgram, err := tea.NewProgram(model,
    tea.WithAltScreen(),
    tea.WithOutput(tty),
    tea.WithInput(tty),
).Run()
```

**Why**: Without `/dev/tty`, lipgloss detects stdout (pipe) and strips all colors/styles. `OpenTTY` also calls `lipgloss.SetColorProfile(termenv.NewOutput(tty).ColorProfile())` to restore styles globally.

### Pattern 3: CommandDef flags — decision tree

| Flag | Use when | Example |
|------|----------|---------|
| `CdOutput=true` | Command prints a path → shell must `cd` into it | `gp` |
| `NeedsEval=true` | Script may need `export`/`PATH` in parent shell | `gpd` |
| `HasCompletions=true` | Command takes project name as arg | `gp`, `gpd` |

**Rule**: Only `gpd` uses `NeedsEval`. Any new command wanting eval must justify it explicitly — it's a security surface.

Generated wrappers per flag combination:

```bash
# CdOutput=true (gp)
gp() {
  local targetDir
  targetDir="$(arielsurco-cli project go "$@")"
  [ -n "$targetDir" ] && cd "$targetDir"
}

# NeedsEval=true (gpd)
gpd() {
  eval "$(arielsurco-cli project dev "$@")"
}

# Neither (gpa, gpr)
gpa() {
  arielsurco-cli project add "$@"
}
```

### Pattern 4: Tab Completion Wiring

Completions are emitted by `shell-init` for commands with `HasCompletions=true`.
They call Cobra's `__complete` subcommand at TAB-time — always fresh, never cached.

```bash
# bash
_gp_completions() {
  local currentWord="${COMP_WORDS[COMP_CWORD]}"
  COMPREPLY=($(arielsurco-cli __complete project go "$currentWord" 2>/dev/null | grep -v '^:'))
}
complete -F _gp_completions gp

# zsh
_gp() {
  local -a completionItems
  completionItems=($(arielsurco-cli __complete project go "${words[2]}" 2>/dev/null | grep -v '^:'))
  _describe 'project' completionItems
}
compdef _gp gp
```

**After changing the Registry**: always regenerate golden files:
```bash
go test ./internal/shell/... -update
```

### Pattern 5: Cancel = silent exit

All TUI commands return `nil` on Escape/Ctrl+C — no error, no message:

```go
selectionResult := finalProgram.(projectlist.Model).Result()
if selectionResult.Cancelled {
    return nil  // ✅ silent — not fmt.Errorf("cancelled")
}
```
