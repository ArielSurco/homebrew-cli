# AGENTS.md — arielsurco-cli

Guía de entrada para agentes y contribuidores. Leé las skills relevantes antes de tocar código.

## Skills disponibles

| Skill | Cuándo leerla |
|-------|--------------|
| [go-testing](.agents/skills/go-testing/SKILL.md) | Al escribir o modificar cualquier test |
| [shell-integration](.agents/skills/shell-integration/SKILL.md) | Al tocar wrappers, TTY detection, shell-init, eval rules |
| [tui-patterns](.agents/skills/tui-patterns/SKILL.md) | Al crear o modificar modelos Bubbletea |
| [module-extension](.agents/skills/module-extension/SKILL.md) | Al agregar un comando o módulo nuevo |
| [config-patterns](.agents/skills/config-patterns/SKILL.md) | Al tocar config, XDG, atomic save, Service layer |
| [release](.agents/skills/release/SKILL.md) | Al preparar un release o modificar CI |

## Stack rápido

Go + Cobra + Fang v2 + Bubbletea v1 + bubbles/list + Lip Gloss + TOML + XDG

## Reglas no negociables

- **TDD**: test rojo antes de implementar. Sin excepciones.
- **Linter**: `golangci-lint run` con 0 issues antes de cada commit.
- **Commits atómicos**: conventional commits, un commit por feature/fix.
- **Naming**: sin variables de una letra. Ver skill go-testing.

## Comandos

| Alias | Cobra | Descripción |
|-------|-------|-------------|
| `gp [name]` | `project go` | Navegar a proyecto |
| `gpd [name]` | `project dev` | Correr dev script |
| `gpa <name> <path>` | `project add` | Registrar proyecto |
| `gpr [name]` | `project remove` | Eliminar proyecto |

## Backlog post v0.1.0

- Alias customization TUI (`[aliases]` en active.toml)
- `gpd` exact-match shortcut (igual que `gp`)
- Módulos futuros: opensearch (`o` prefix), sentry (`s` prefix)
