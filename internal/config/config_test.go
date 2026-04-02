package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ArielSurco/cli/internal/config"
)

func TestLoad_MissingFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil Config")
	}
	if len(cfg.Projects) != 0 {
		t.Errorf("expected empty projects, got %d", len(cfg.Projects))
	}
}

func TestLoad_ValidTOML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "arielsurco-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tomlContent := `[[projects]]
name = "myapp"
path = "/home/user/myapp"
dev_script = "npm run dev"
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(tomlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(cfg.Projects))
	}
	loadedProject := cfg.Projects[0]
	if loadedProject.Name != "myapp" {
		t.Errorf("name: got %q, want %q", loadedProject.Name, "myapp")
	}
	if loadedProject.Path != "/home/user/myapp" {
		t.Errorf("path: got %q, want %q", loadedProject.Path, "/home/user/myapp")
	}
	if loadedProject.DevScript != "npm run dev" {
		t.Errorf("dev_script: got %q, want %q", loadedProject.DevScript, "npm run dev")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	original := &config.Config{
		Projects: []config.Project{
			{Name: "proj1", Path: "/path/one", DevScript: "make dev"},
			{Name: "proj2", Path: "/path/two"},
		},
	}

	if err := config.Save(original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Projects) != len(original.Projects) {
		t.Fatalf("project count: got %d, want %d", len(loaded.Projects), len(original.Projects))
	}
	for index, expectedProject := range original.Projects {
		loadedProject := loaded.Projects[index]
		if loadedProject.Name != expectedProject.Name || loadedProject.Path != expectedProject.Path || loadedProject.DevScript != expectedProject.DevScript {
			t.Errorf("project[%d]: got %+v, want %+v", index, loadedProject, expectedProject)
		}
	}
}

func TestLoad_WithLocalConfig_MergesDevScript(t *testing.T) {
	xdgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	// Write global config
	cfgDir := filepath.Join(xdgDir, "arielsurco-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	globalToml := `[[projects]]
name = "webapp"
path = "/home/user/webapp"
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(globalToml), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write local config in a temp CWD
	cwdDir := t.TempDir()
	localToml := `[project]
name = "webapp"
dev_script = "yarn dev"
`
	if err := os.WriteFile(filepath.Join(cwdDir, ".arielsurco-cli.toml"), []byte(localToml), 0o644); err != nil {
		t.Fatal(err)
	}

	// Change CWD
	origDir, _ := os.Getwd()
	if err := os.Chdir(cwdDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(cfg.Projects))
	}
	if cfg.Projects[0].DevScript != "yarn dev" {
		t.Errorf("dev_script: got %q, want %q", cfg.Projects[0].DevScript, "yarn dev")
	}
}

func TestLoad_WithLocalConfig_NoMatchingProject(t *testing.T) {
	xdgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	cfgDir := filepath.Join(xdgDir, "arielsurco-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	globalToml := `[[projects]]
name = "webapp"
path = "/home/user/webapp"
dev_script = "original-script"
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(globalToml), 0o644); err != nil {
		t.Fatal(err)
	}

	cwdDir := t.TempDir()
	localToml := `[project]
name = "other-project"
dev_script = "yarn dev"
`
	if err := os.WriteFile(filepath.Join(cwdDir, ".arielsurco-cli.toml"), []byte(localToml), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(cwdDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(cfg.Projects))
	}
	if cfg.Projects[0].DevScript != "original-script" {
		t.Errorf("dev_script should not change, got %q", cfg.Projects[0].DevScript)
	}
}

func TestLoadActive_MissingFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	am, err := config.LoadActive()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if am == nil {
		t.Fatal("expected non-nil ActiveModules")
	}
	if len(am.Modules.Active) != 0 {
		t.Errorf("expected empty active modules, got %d", len(am.Modules.Active))
	}
}

func TestSaveAndLoadActive_RoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	original := &config.ActiveModules{
		Modules: config.ModulesSection{
			Active: []string{"git", "node", "docker"},
		},
	}

	if err := config.SaveActive(original); err != nil {
		t.Fatalf("SaveActive failed: %v", err)
	}

	loaded, err := config.LoadActive()
	if err != nil {
		t.Fatalf("LoadActive failed: %v", err)
	}

	if len(loaded.Modules.Active) != len(original.Modules.Active) {
		t.Fatalf("active count: got %d, want %d", len(loaded.Modules.Active), len(original.Modules.Active))
	}
	for index, moduleName := range original.Modules.Active {
		if loaded.Modules.Active[index] != moduleName {
			t.Errorf("active[%d]: got %q, want %q", index, loaded.Modules.Active[index], moduleName)
		}
	}
}
