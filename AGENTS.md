# AGENTS.md — arielsurco-cli

Guía de decisiones de diseño y convenciones para cualquier agente o persona que contribuya a este proyecto. Leé esto antes de escribir una sola línea de código.

---

## Stack

| Componente | Decisión |
|------------|---------|
| Lenguaje | Go (latest stable) |
| CLI framework | Cobra + Fang v2 (`charm.land/fang/v2`) |
| TUI | Bubbletea v1 stable + `bubbles/list` + `bubbles/textinput` |
| Styling | Lip Gloss |
| Config format | TOML via `github.com/pelletier/go-toml/v2` |
| Config paths | XDG via `github.com/adrg/xdg` — **nunca hardcodear `~/.config`** |
| Linter | golangci-lint (`default: standard` + revive, misspell, nolintlint, gocritic) |
| Testing TUI | `github.com/charmbracelet/x/exp/teatest` |
| Fuzzy match | `github.com/sahilm/fuzzy` |
| Terminal detect | `golang.org/x/term` + `github.com/muesli/termenv` |

---

## Reglas no negociables

### TDD — siempre
Escribir el test que falla primero, luego implementar. Sin excepciones. Cada feature nueva empieza con un test rojo.

### Linter limpio — siempre
`golangci-lint run` debe pasar con 0 issues antes de cualquier commit.

### Commits atómicos por fase/feature
Un commit por feature o fix. Conventional commits: `feat:`, `fix:`, `refactor:`, `chore:`, `test:`. Sin "Co-Authored-By" ni atribución de IA.

### Nombres descriptivos — sin variables de una letra
- ❌ `p`, `s`, `v`, `i`, `m`, `tmp`, `lc`, `am`
- ✅ `foundProject`, `existingProject`, `index`, `activeModules`, `tempFile`, `override`
- Receivers: `svc` no `s`, descriptivos siempre
- Variables de test: `shellCommand` no `cmd`, `expectedCommand`, `loadedProject`

---

## Arquitectura del proyecto

```
cmd/                        ← Solo wiring de Cobra. Cero lógica de negocio.
  root.go                   ← Fang v2 wiring, Version var
  setup.go                  ← setup TUI command
  shellinit.go              ← shell-init command
  project/
    go.go                   ← gp — smart navigation
    dev.go                  ← gpd — dev script via eval
    add.go                  ← gpa — register project
    remove.go               ← gpr — mirrors gp behavior
internal/
  config/                   ← Leaf package, zero deps internas
  project/                  ← Service + sentinel errors (deps: config)
  module/                   ← Registry compilado, zero deps
  shell/                    ← TTY detection + Generate() (deps: module)
  tui/
    projectlist/            ← Bubbletea list (deps: config)
    setup/                  ← Bubbletea checklist (deps: module)
```

**Regla de dependencias (DAG, sin ciclos):**
```
config   module          ← leaf packages
   ↑        ↑
project   shell
   ↑        ↑
tui/*   tui/setup
   ↑        ↑
cmd/*                    ← única capa que persiste y maneja exit codes
```

---

## Patrones de implementación

### Service layer
- `Service` opera en memoria sobre `*config.Config`
- **Service NUNCA llama `config.Save()`** — eso es responsabilidad del cmd layer
- Sentinel errors en `internal/project/errors.go`
- El cmd layer wrappea los errores con contexto user-friendly

### Config
- `Load()` retorna `&Config{}, nil` si no existe el archivo — **no es un error**
- Save es siempre atómico: `os.CreateTemp(dir, "*.toml.tmp")` + `os.Rename`
- `xdg.Reload()` antes de resolver paths (macOS cachea al init)
- Config local `.arielsurco-cli.toml` en CWD tiene precedencia sobre global para `dev_script`
- `filepath.Abs()` al guardar paths — acepta relativos, guarda absolutos

### TUI models
- **Los modelos NUNCA hacen I/O** — datos inyectados en el constructor
- Resultado leído de `FinalModel(t).(ModelType).Result()` después de que el programa termina
- Cancel (Escape/Ctrl+C) retorna **nil** — salida silenciosa, sin error visible al usuario
- TUI se lanza con `/dev/tty` para I/O correcto dentro de `$()`:

```go
tty, err := shell.OpenTTY()   // abre /dev/tty + configura lipgloss color profile
defer tty.Close()             //nolint:errcheck
tea.NewProgram(model,
    tea.WithAltScreen(),
    tea.WithOutput(tty),
    tea.WithInput(tty),
).Run()
```

### TTY detection — CRÍTICO
```go
// Para decidir si lanzar TUI: usar STDIN (no stdout)
shell.IsInteractiveSession()   // term.IsTerminal(os.Stdin.Fd())
// ↑ Funciona dentro de $() donde stdout es pipe pero stdin sigue siendo terminal
```
**Nunca usar `IsTerminal()` (stdout) para decisiones de TUI** — dentro de `$()` stdout es un pipe.

### Comportamiento smart de comandos con TUI (gp, gpd, gpr)
```
TTY + sin args           → TUI vacía
TTY + arg exacto         → acción directa sin TUI
TTY + arg con fuzzy match → TUI con preFilter aplicado
TTY + arg sin matches    → error "not found"
non-TTY + sin args       → error ErrNameRequired
non-TTY + arg exacto     → acción directa, output a stdout
```

---

## Sistema de módulos y shell wrappers

### CommandDef
```go
type CommandDef struct {
    Alias          string  // "gp"
    CobraCmd       string  // "project go"
    NeedsEval      bool    // true = eval "$(...)" en el wrapper
    CdOutput       bool    // true = captura output y hace cd
    HasCompletions bool    // true = emite completion wiring en shell-init
    Description    string
}
```

### Cuándo usar cada flag
| Flag | Cuándo usarlo | Ejemplo |
|------|---------------|---------|
| `CdOutput=true` | El comando imprime un path para navegar | `gp` |
| `NeedsEval=true` | El script puede necesitar `export`/`PATH` en el padre | `gpd` |
| `HasCompletions=true` | El comando acepta nombres de proyectos como argumento | `gp`, `gpd` |

**Regla**: solo `gpd` usa `eval`. Cualquier comando nuevo debe justificar explícitamente el uso de eval.

### Agregar un nuevo módulo/comando
1. Agregar `Module` entry en `internal/module/registry.go`
2. Agregar Cobra command bajo `cmd/{module}/`
3. Si el comando acepta nombres de proyectos: `ValidArgsFunction: completeProjectNames` + `HasCompletions: true`
4. Correr `go test ./internal/shell/... -update` para regenerar golden files
5. `shell-init`, `setup` y los wrappers se actualizan automáticamente

### Naming de aliases
Patrón: `{inicial-módulo}{inicial-subcomando}{inicial-comando}`
- `gp` = **g**o **p**roject
- `gpd` = **g**o **p**roject **d**ev
- Opensearch futuro: prefijo `o` (ej: `opi`, `oq`)
- Sentry futuro: prefijo `s` (resolver colisiones al definir)

---

## Testing

### Reglas
```go
// SIEMPRE table-driven tests
tests := []struct {
    name string
    ...
}{...}
for _, testCase := range tests {
    t.Run(testCase.name, func(t *testing.T) {...})
}

// SIEMPRE aislar config con XDG
t.Setenv("XDG_CONFIG_HOME", t.TempDir())

// CRÍTICO: cualquier test que llame config.Save() transitivamente
// DEBE tener el setenv de arriba — si no, escribe al config real del usuario
```

### TUI tests con teatest
```go
tuiModel := projectlist.New(projects, "")  // inyectar datos, nunca cargar config adentro
testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))
testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
finalModel := testModel.FinalModel(t).(projectlist.Model)
result := finalModel.Result()
```

### Golden files para shell.Generate()
- Archivos en `internal/shell/testdata/golden_bash.txt` y `golden_zsh.txt`
- Actualizar con: `go test ./internal/shell/... -update`
- Siempre correr el update después de cambiar `CommandDef` en el registry

---

## Homebrew / Release

- **GoReleaser key**: `homebrew_casks` (NO `brews` — deprecado en v2)
- **Tap**: mismo repo `ArielSurco/cli` — no repo separado
- **Token**: `GITHUB_TOKEN` estándar de GitHub Actions — no se necesita PAT extra
- **Install para usuarios**:
  ```bash
  brew tap ArielSurco/cli https://github.com/ArielSurco/cli
  brew install arielsurco-cli
  ```
- **Release**: `git tag vX.Y.Z && git push origin vX.Y.Z`
- **4 builds**: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64

---

## MCP-readiness

El binary está diseñado para ser consumible por MCP en el futuro sin cambios:
- Non-TTY mode siempre imprime output limpio a stdout
- Errores van a stderr
- Sin decoraciones ANSI en modo non-TTY
- Output parseable (paths crudos, sin formato adicional)
- Patrón `--json` puede agregarse a comandos cuando MCP lo requiera

---

## Backlog (post v0.1.0)

| Feature | Descripción | Complejidad |
|---------|-------------|-------------|
| Alias customization TUI | `[aliases]` en active.toml, textinput por comando en setup | Media |
| `gpd` exact-match shortcut | TTY + arg exacto → correr dev script directo sin TUI | Baja |
| `gpr` tab completion en alias | `HasCompletions=true` para gpr | Baja |
| Módulo opensearch | Prefijo `o`, comandos a definir | Variable |
| Módulo sentry | Prefijo `s`, comandos a definir | Variable |

---

## Comandos disponibles

| Alias | Cobra | Descripción |
|-------|-------|-------------|
| `gp [name]` | `project go` | Navegar a proyecto (smart: exact→cd directo, fuzzy→TUI filtrada) |
| `gpd [name]` | `project dev` | Correr dev script via eval |
| `gpa <name> <path>` | `project add` | Registrar proyecto (acepta paths relativos) |
| `gpr [name]` | `project remove` | Eliminar proyecto (mismo comportamiento que gp) |
| — | `setup` | Seleccionar módulos activos (TUI) |
| — | `shell-init` | Emitir wrappers para shell actual |

**Setup inicial para el usuario:**
```bash
arielsurco-cli setup
# Agregar a ~/.zshrc o ~/.bashrc:
eval "$(arielsurco-cli shell-init)"
```
