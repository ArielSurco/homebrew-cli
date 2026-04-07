---
name: shell-integration
description: >
  Shell wrapper generation, TTY detection, eval rules, shell-init injection, and OpenTTY pattern
  for arielsurco-cli.
  Trigger: When modifying shell-init, wrappers, CommandDef, TTY detection, setup auto-injection,
  or adding commands that interact with the parent shell.
metadata:
  author: arielsurco
  version: "2.0"
---

## When to Use

- Modifying `internal/shell/detect.go`, `generate.go`, or `inject.go`
- Adding a new command that needs a shell alias
- Changing how wrappers handle cd, eval, or completions
- Modifying the setup command's shell-init injection
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

### Pattern 6: User-facing messages MUST go to stderr in CdOutput commands

Commands with `CdOutput: true` capture stdout inside `$()` and pass it to `cd`.
Any `fmt.Printf` or `fmt.Println` to stdout contaminates the captured path.

```go
// ❌ WRONG — "Project removed." ends up inside cd "$targetDir"
fmt.Printf("Project %q removed.\n", name)

// ✅ CORRECT — stderr is never captured by the shell wrapper
fmt.Fprintf(os.Stderr, "Project %q removed.\n", name)
```

**Rule**: In any function that can be called from a `CdOutput` command (directly or transitively), ALL user-facing messages MUST use `fmt.Fprintf(os.Stderr, ...)`. This includes confirmation prompts, success messages, and info messages.

Affected commands today: `gp` (`project go`). Any future `CdOutput` command inherits this rule.

### Pattern 7: Shell-init auto-injection (setup command)

The `setup` command automatically injects a source block into `~/.bashrc` or `~/.zshrc`.

#### Fixed path for shell-init.sh

```
~/.config/arielsurco-cli/shell-init.sh
```

**DO NOT use XDG** (`xdg.ConfigHome`) for the shell-init file path. On macOS, XDG resolves
to `~/Library/Application Support/` which has spaces and causes shell parsing issues.
The fixed `~/.config/` path works identically on macOS and Linux.

Note: The main config (`config.toml`, `active-modules.toml`) still uses XDG via `config.go`.
Only the shell-init file uses the fixed path.

#### The injected block uses `$HOME`, not `~`

```bash
# ✅ CORRECT — $HOME expands inside double quotes
if [[ -f $HOME/.config/arielsurco-cli/shell-init.sh ]]; then
  source $HOME/.config/arielsurco-cli/shell-init.sh

# ❌ WRONG — ~ does NOT expand inside double quotes
if [[ -f "~/.config/arielsurco-cli/shell-init.sh" ]]; then
```

#### Injection is idempotent

`InjectShellInit` checks for the marker `arielsurco-cli/shell-init.sh` in the rc file.
If found, it skips injection and returns `false`. Running setup multiple times is safe.

#### Test isolation: always pass homeDir

Tests MUST use `InjectShellInitWithHome(shell, homeDir)` with a `t.TempDir()` home directory.
Using `InjectShellInit(shell)` in tests writes to the real `~/.zshrc`.

```go
// ✅ CORRECT — isolated
homeDir := t.TempDir()
shell.InjectShellInitWithHome(shell.Zsh, homeDir)

// ❌ WRONG — pollutes real ~/.zshrc
shell.InjectShellInit(shell.Zsh)
```

Similarly, `cmd.RunSetupWithResult` accepts a `homeDir` parameter for test isolation:

```go
// ✅ CORRECT
cmd.RunSetupWithResult(modules, true, t.TempDir())

// ❌ WRONG
cmd.RunSetupWithResult(modules, true, "")
```
