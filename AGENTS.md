# AGENTS.md — arielsurco-cli

Read the relevant skills before writing any code.

## Skills

| Skill | When to load |
|-------|-------------|
| [go-testing](.agents/skills/go-testing/SKILL.md) | Writing or modifying any test |
| [shell-integration](.agents/skills/shell-integration/SKILL.md) | Shell wrappers, TTY detection, shell-init, eval rules |
| [tui-patterns](.agents/skills/tui-patterns/SKILL.md) | Creating or modifying Bubbletea models |
| [module-extension](.agents/skills/module-extension/SKILL.md) | Adding a new command or module |
| [config-patterns](.agents/skills/config-patterns/SKILL.md) | Config, XDG, atomic save, Service layer |
| [release](.agents/skills/release/SKILL.md) | Preparing a release or modifying CI |

## Stack

Go + Cobra + Fang v2 + Bubbletea v1 + bubbles/list + Lip Gloss + TOML + XDG

## Non-negotiable rules

- **TDD**: failing test before implementation. No exceptions.
- **Linter**: `golangci-lint run` must report 0 issues before every commit.
- **Atomic commits**: conventional commits, one commit per feature/fix.
- **Naming**: no single-letter variables. See go-testing skill.

## Commands

| Alias | Cobra | Description |
|-------|-------|-------------|
| `gp [name]` | `project go` | Navigate to a project |
| `gpd [name]` | `project dev` | Run the project dev script |
| `gpa <name> <path>` | `project add` | Register a project |
| `gpr [name]` | `project remove` | Unregister a project |
