---
name: tui-patterns
description: >
  Bubbletea model patterns for arielsurco-cli: data injection, /dev/tty, fixed height,
  cancel behavior, and Fang styling.
  Trigger: When creating or modifying Bubbletea TUI models or launching tea.NewProgram.
metadata:
  author: arielsurco
  version: "1.0"
---

## When to Use

- Creating a new Bubbletea model in `internal/tui/`
- Modifying `projectlist` or `setup` models
- Launching `tea.NewProgram` from a command
- Debugging missing colors or pagination issues

---

## Critical Patterns

### Pattern 1: Models never do I/O — inject data at construction

```go
// ✅ CORRECT
tuiModel := projectlist.New(cfg.Projects, preFilter)  // data injected

// ❌ WRONG — model loading config internally breaks testability
func (model Model) Init() tea.Cmd {
    cfg, _ := config.Load()  // never do this inside a model
    ...
}
```

Result is read after program exits:
```go
finalProgram, err := tea.NewProgram(tuiModel, ...).Run()
result := finalProgram.(projectlist.Model).Result()
```

### Pattern 2: Launch pattern — always use /dev/tty

```go
tty, err := shell.OpenTTY()
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

Never use `tea.WithOutput(os.Stderr)` — it doesn't fix lipgloss color detection.

### Pattern 3: Fixed height for lists

`bubbles/list` has a circular pagination bug when height is computed dynamically.
Use a fixed height large enough for the max visible items:

```go
const (
    listWidth  = 80
    listHeight = 20  // fits up to 12 items without pagination
)

listModel := list.New(items, delegate, listWidth, listHeight)
listModel.SetShowStatusBar(false)
```

On `WindowSizeMsg`, update width only — never height:
```go
if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
    model.list.SetSize(sizeMsg.Width, model.list.Height())  // preserve height
}
```

### Pattern 4: Compact list delegate

```go
compactDelegate := list.NewDefaultDelegate()
compactDelegate.ShowDescription = false  // single line per item
compactDelegate.SetSpacing(0)            // no gap between items
```

### Pattern 5: Cancel returns nil

```go
func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if keyMsg, ok := msg.(tea.KeyMsg); ok {
        switch keyMsg.Type {
        case tea.KeyEnter:
            selected := model.list.SelectedItem()
            if selected != nil {
                model.selectedItem = &selected
                return model, tea.Quit
            }
        case tea.KeyCtrlC, tea.KeyEsc:
            model.cancelled = true
            return model, tea.Quit  // Cmd layer checks Cancelled and returns nil
        }
    }
    ...
}
```

### Pattern 6: Fang v2 styling — remove codeblock box

In `cmd/root.go`, Fang is configured with `color.Transparent` on Codeblock to remove the background box from the usage line:

```go
fang.Execute(context.Background(), rootCmd,
    fang.WithColorSchemeFunc(func(lightDark fanglipgloss.LightDarkFunc) fang.ColorScheme {
        scheme := fang.DefaultColorScheme(lightDark)
        scheme.Codeblock = color.Transparent
        return scheme
    }),
)
```

Do not remove Fang — it provides command/flag coloring throughout help output.

### Pattern 7: Guard custom key handlers against bubbles/list filter mode

`bubbles/list` intercepts keypresses when the user is actively typing a filter (`/` then text).
Custom key handlers (`d`, `e`, etc.) MUST check filter state before acting — otherwise the key
is treated as an action instead of filter input.

```go
func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if keyMsg, ok := msg.(tea.KeyMsg); ok {
        // ✅ ALWAYS check this first
        if model.list.FilterState() == list.Filtering {
            updatedList, listCmd := model.list.Update(msg)
            model.list = updatedList
            return model, listCmd
        }

        switch string(keyMsg.Runes) {
        case "d":
            // safe to handle — not in filter mode
        }
    }
    ...
}
```

Also: in confirm modes (e.g. `modeConfirmDelete`), do NOT forward key messages to
`model.list.Update()` — the list should be frozen while waiting for confirmation.

### Pattern 8: Separate constructor for destructive TUI contexts

When a TUI is opened from a destructive command (e.g. `gpr`), use a dedicated constructor
that sets the initial state to communicate intent clearly — different title, warning-colored
footer, and Enter triggers confirmation instead of navigation.

```go
// ✅ For gpr — user sees "Remove a project" + red footer from the start
tuiModel := projectlist.NewForDelete(cfg.Projects, preFilter)

// ✅ For gp — default navigate context
tuiModel := projectlist.New(cfg.Projects, preFilter)
```

The `NewForDelete` constructor sets `deleteMode: true` on the model, which changes:
- `listModel.Title` → `"Remove a project"`
- Footer in `modeNavigate` → coral-colored `"select project to remove — [enter] confirm  [esc] cancel"`
- Enter → transitions to `modeConfirmDelete` instead of returning `ActionNavigate`

Apply this pattern whenever the same TUI component is reused across commands with different intents.
