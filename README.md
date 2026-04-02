# arielsurco-cli

A personal developer CLI for fast project navigation and development workflow automation. It lets you jump between registered projects, run their dev scripts, and manage your project registry — all from a short shell alias.

## Install

```bash
brew tap ArielSurco/cli https://github.com/ArielSurco/cli
brew install arielsurco-cli
```

## First-time setup

```bash
arielsurco-cli setup
# Add to ~/.zshrc or ~/.bashrc:
eval "$(arielsurco-cli shell-init)"
```

Restart your shell (or `source ~/.zshrc`) after adding the line above.

## Upgrade

```bash
brew upgrade arielsurco-cli
```

## Commands

| Alias | Command | Description |
|-------|---------|-------------|
| `gp [name]` | `project go` | Navigate to a project |
| `gpd [name]` | `project dev` | Run the project's dev script |
| `gpa <name> <path>` | `project add` | Register a project |
| `gpr <name>` | `project remove` | Unregister a project |

All aliases are injected into your shell by `shell-init` and require the eval line in your shell config.

## Configuration

arielsurco-cli stores its configuration in the XDG config directory (typically `~/.config/arielsurco-cli/`).

- **Global config** (`config.toml`): default editor, shell, and global preferences.
- **Local override** (`.arielsurco-cli.toml` in a project root): per-project overrides such as a custom dev script command. Local config takes precedence over global config.

## Tab completion

Shell completions are automatically available after sourcing `shell-init`. No extra steps needed — completions for all commands and registered project names are injected at shell startup.
