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
