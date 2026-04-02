---
name: module-extension
description: >
  Step-by-step guide to add a new command or module to arielsurco-cli while maintaining
  full consistency: registry, Cobra wiring, shell wrappers, completions, TUI, tests.
  Trigger: When adding a new command, alias, or module category to the CLI.
metadata:
  author: arielsurco
  version: "1.0"
---

## When to Use

- Adding a new command to an existing module (e.g. `gpl` for project list)
- Creating a new module category (e.g. opensearch, sentry)
- Adding a new alias that needs shell completion or TUI

---

## Adding a Command to an Existing Module

### Step 1: Add to Registry

`internal/module/registry.go`:

```go
{
    Alias:          "gpl",
    CobraCmd:       "project list",
    NeedsEval:      false,   // only true if script needs export/PATH in parent
    CdOutput:       false,   // only true if command prints a path for cd
    HasCompletions: false,   // true if command takes project name as arg
    Description:    "List all configured projects",
},
```

**Flag decision guide:**

| Flag | Set true when |
|------|--------------|
| `CdOutput` | Binary prints a directory path → shell must `cd` into it |
| `NeedsEval` | Output may contain `export` or `PATH` changes for parent shell |
| `HasCompletions` | Command accepts a project name as argument |

### Step 2: Create Cobra command

`cmd/project/list.go`:

```go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all configured projects",
    RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("loading config: %w", err)
    }
    svc := project.NewService(cfg)
    for _, existingProject := range svc.List() {
        fmt.Printf("%-20s %s\n", existingProject.Name, existingProject.Path)
    }
    return nil
}
```

Register in `cmd/project/project.go`:
```go
func init() {
    Cmd.AddCommand(listCmd)
}
```

### Step 3: Update golden files

```bash
go test ./internal/shell/... -update
```

### Step 4: Tests (TDD — write before implementing)

```go
func TestListCommand_PrintsProjects(t *testing.T) {
    cfg := &config.Config{
        Projects: []config.Project{
            {Name: "my-app", Path: "/projects/my-app"},
        },
    }
    var outputBuffer bytes.Buffer
    // test the command output...
}
```

---

## Adding a New Module

Modules group related commands under a prefix (e.g. `o` for opensearch).

### Naming convention

`{module-initial}{subcommand-initial}{command-initial}`

Examples:
- `o` prefix → opensearch module
- `s` prefix → sentry module
- Check for conflicts with existing shell commands before committing to a prefix

### Step 1: Add Module to Registry

```go
{
    Name: "opensearch",
    Commands: []CommandDef{
        {Alias: "oq",  CobraCmd: "opensearch query",  NeedsEval: false, HasCompletions: false, Description: "Query an index"},
        {Alias: "oi",  CobraCmd: "opensearch index",  NeedsEval: false, HasCompletions: false, Description: "List indices"},
    },
},
```

### Step 2: Create `cmd/opensearch/` with Cobra commands

Follow the same structure as `cmd/project/`:
- `opensearch.go` — group command (`Cmd`)
- `query.go`, `index.go` — individual commands
- Add `Cmd` to `rootCmd` in `cmd/root.go`

### Step 3: Update setup TUI (automatic)

`setup` reads from `module.Registry` — new module appears automatically in the TUI checklist.

### Step 4: Update golden files and test

```bash
go test ./internal/shell/... -update
go test ./...
golangci-lint run
```

---

## Smart Navigation Pattern (mirrors gp/gpr)

Any command that selects from a list should use this pattern:

```go
// TTY + exact match → direct action
// TTY + fuzzy match → TUI with preFilter
// TTY + no match → error
// non-TTY + no args → ErrNameRequired
// non-TTY + arg → direct lookup

if projectName != "" {
    if err := doActionByName(projectName, cfg); err == nil {
        return nil  // exact match handled
    } else if !errors.Is(err, project.ErrNotFound) {
        return err
    }
    // fuzzy check
    projectNames := make([]string, len(cfg.Projects))
    for index, existingProject := range cfg.Projects {
        projectNames[index] = existingProject.Name
    }
    if len(fuzzy.Find(projectName, projectNames)) == 0 {
        return fmt.Errorf("project %q not found", projectName)
    }
}
// fall through to TUI
```
