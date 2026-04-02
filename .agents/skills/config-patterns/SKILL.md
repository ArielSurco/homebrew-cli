---
name: config-patterns
description: >
  Config load/save patterns for arielsurco-cli: XDG paths, atomic write, local override,
  path resolution, and Service layer rules.
  Trigger: When modifying config loading/saving, adding new config fields, or touching
  internal/config or internal/project.
metadata:
  author: arielsurco
  version: "1.0"
---

## When to Use

- Modifying `internal/config/config.go`
- Adding new fields to `Config`, `Project`, or `ActiveModules`
- Implementing a command that reads or writes config
- Working with `internal/project/service.go`

---

## Critical Patterns

### Pattern 1: XDG paths — never hardcode

```go
// ✅ CORRECT
import "github.com/adrg/xdg"

func GlobalConfigPath() (string, error) {
    xdg.Reload()  // required on macOS — library caches at init time
    return filepath.Join(xdg.ConfigHome, "arielsurco-cli", "config.toml"), nil
}

// ❌ WRONG
path := filepath.Join(os.Getenv("HOME"), ".config", "arielsurco-cli", "config.toml")
```

`xdg.Reload()` is mandatory before every path resolution — without it `t.Setenv("XDG_CONFIG_HOME")` has no effect in tests.

### Pattern 2: Atomic save

```go
func atomicSaveTOML(path string, value any) error {
    configDir := filepath.Dir(path)
    if err := os.MkdirAll(configDir, 0o755); err != nil {
        return err
    }
    data, err := toml.Marshal(value)
    if err != nil {
        return err
    }
    tempFile, err := os.CreateTemp(configDir, "*.toml.tmp")
    if err != nil {
        return err
    }
    tempFilePath := tempFile.Name()
    if _, err := tempFile.Write(data); err != nil {
        _ = tempFile.Close()
        _ = os.Remove(tempFilePath)
        return err
    }
    if err := tempFile.Close(); err != nil {
        _ = os.Remove(tempFilePath)
        return err
    }
    return os.Rename(tempFilePath, path)  // POSIX atomic
}
```

**Always use `os.CreateTemp` in the SAME directory as the target** — cross-device rename is not atomic.

### Pattern 3: Load returns empty config when file missing

```go
func Load() (*Config, error) {
    globalPath, _ := GlobalConfigPath()
    data, err := os.ReadFile(globalPath)
    if errors.Is(err, os.ErrNotExist) {
        return &Config{}, nil  // ✅ not an error — first run
    }
    ...
}
```

### Pattern 4: Local config override

`.arielsurco-cli.toml` in CWD overrides `dev_script` for matching project name:

```go
// merge: local wins for dev_script, match by name
for index := range cfg.Projects {
    if cfg.Projects[index].Name == override.Project.Name && override.Project.DevScript != "" {
        cfg.Projects[index].DevScript = override.Project.DevScript
        break
    }
}
```

### Pattern 5: Path resolution on add

Always resolve to absolute at write time:

```go
absolutePath, err := filepath.Abs(path)  // accepts ".", "../other", or absolute
```

### Pattern 6: Service layer rules

```go
// ✅ Service operates on *config.Config in memory
// ✅ Service returns sentinel errors (never user-facing strings)
// ❌ Service NEVER calls config.Save() — that's the cmd layer's job

func (svc *Service) Add(name, path, devScript string) error {
    absolutePath, _ := filepath.Abs(path)
    for _, existingProject := range svc.cfg.Projects {
        if existingProject.Name == name {
            return ErrDuplicateName  // sentinel — cmd layer wraps with context
        }
    }
    svc.cfg.Projects = append(svc.cfg.Projects, config.Project{...})
    return nil
    // caller does: config.Save(cfg)
}
```

### Pattern 7: Empty config message

When `cfg.Projects` is empty, show actionable instructions:

```go
if len(cfg.Projects) == 0 {
    return fmt.Errorf("no projects configured\n\nAdd your first project:\n  gpa <name> <path>\n\nExample:\n  gpa my-app /Users/%s/projects/my-app", os.Getenv("USER"))
}
```
