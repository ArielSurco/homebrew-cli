---
name: go-testing
description: >
  Testing patterns for arielsurco-cli: table-driven tests, XDG isolation, teatest for TUI,
  golden files for shell output, and config save isolation.
  Trigger: When writing or modifying any Go test in this project.
metadata:
  author: arielsurco
  version: "1.0"
---

## When to Use

- Writing unit tests for any package
- Testing Bubbletea TUI models with teatest
- Adding golden file tests for shell output
- Testing commands that involve config load/save

---

## Critical Patterns

### Pattern 1: Table-Driven Tests

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    string
        wantErr     bool
    }{
        {name: "valid case", input: "hello", expected: "world", wantErr: false},
        {name: "error case", input: "", expected: "", wantErr: true},
    }
    for _, testCase := range tests {
        t.Run(testCase.name, func(t *testing.T) {
            result, err := DoSomething(testCase.input)
            if (err != nil) != testCase.wantErr {
                t.Errorf("error = %v, wantErr %v", err, testCase.wantErr)
            }
            if result != testCase.expected {
                t.Errorf("got %q, want %q", result, testCase.expected)
            }
        })
    }
}
```

### Pattern 2: XDG Isolation (CRITICAL)

**Any test that transitively calls `config.Save()` or `config.SaveActive()` MUST have this:**

```go
func TestSomething(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", t.TempDir())
    // ...
}
```

Without this, the test writes to the user's real `~/.config/arielsurco-cli/config.toml`.
Use `t.TempDir()` — cleaned up automatically after the test.

### Pattern 3: TUI Tests with teatest

Models never load config internally — inject data at construction:

```go
func TestProjectList_SelectWithEnter(t *testing.T) {
    projects := []config.Project{
        {Name: "my-app", Path: "/projects/my-app"},
        {Name: "api",    Path: "/projects/api"},
    }
    tuiModel := projectlist.New(projects, "")
    testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

    teatest.WaitFor(t, testModel.Output(),
        func(output []byte) bool { return bytes.Contains(output, []byte("my-app")) },
        teatest.WithDuration(3*time.Second),
    )

    testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
    testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

    finalModel := testModel.FinalModel(t).(projectlist.Model)
    selectionResult := finalModel.Result()
    if selectionResult.Cancelled {
        t.Error("expected not cancelled")
    }
    if selectionResult.Project.Name != "my-app" {
        t.Errorf("got %q, want %q", selectionResult.Project.Name, "my-app")
    }
}
```

### Pattern 4: Golden File Tests (shell output)

```go
var updateGolden = flag.Bool("update", false, "update golden files")

func TestGenerate_Bash(t *testing.T) {
    output := shell.Generate(module.Registry, shell.Bash)
    goldenPath := filepath.Join("testdata", "golden_bash.txt")
    if *updateGolden {
        if err := os.WriteFile(goldenPath, []byte(output), 0o644); err != nil {
            t.Fatal(err)
        }
        return
    }
    expected, err := os.ReadFile(goldenPath)
    if err != nil {
        t.Fatalf("golden file missing — run with -update to create: %v", err)
    }
    if string(expected) != output {
        t.Errorf("output does not match golden file\ngot:\n%s\nwant:\n%s", output, expected)
    }
}
```

Update golden files after any Registry change:
```bash
go test ./internal/shell/... -update
```

### Pattern 5: Naming Rules

- ❌ `p`, `s`, `v`, `i`, `m`, `cmd`, `tmp`
- ✅ `foundProject`, `existingProject`, `index`, `shellCommand`, `expectedCommand`
- Test variables: `loadedProject`, `selectionResult`, `testModel`, `finalModel`
- Receivers: `svc` not `s`

### Pattern 6: io.Reader injection for TTY confirmation tests

Functions that read from `/dev/tty` (e.g. `y/N` prompts) are not directly testable.
Split into an inner function accepting `io.Reader` and a production wrapper that opens `/dev/tty`.

```go
// ✅ Inner function — accepts any io.Reader, fully testable
func confirmRemoveFromReader(name string, reader io.Reader) (bool, error) {
    fmt.Fprintf(os.Stderr, "Remove project %q? [y/N]: ", name)
    scanner := bufio.NewScanner(reader)
    if !scanner.Scan() {
        return false, nil
    }
    return strings.ToLower(strings.TrimSpace(scanner.Text())) == "y", nil
}

// ✅ Production wrapper — opens /dev/tty, calls inner function
func confirmRemove(name string) (bool, error) {
    tty, err := shell.OpenTTY()
    if err != nil {
        return false, err
    }
    defer tty.Close() //nolint:errcheck
    return confirmRemoveFromReader(name, tty)
}
```

Export the inner function (e.g. `ConfirmRemoveFromReader`) for use in tests:

```go
// In test
reader := strings.NewReader("y\n")
confirmed, err := cmdproject.ConfirmRemoveFromReader("myapp", reader)
```

Apply this pattern to any function that reads interactive input from the terminal.
