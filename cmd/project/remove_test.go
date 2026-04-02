package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ArielSurco/cli/cmd/project"
	"github.com/pelletier/go-toml/v2"
)

func TestRemoveCommand_Success(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "arielsurco-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := `[[projects]]
name = "myapp"
path = "/path/to/app"

[[projects]]
name = "other"
path = "/path/to/other"
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := project.Cmd
	cmd.SetArgs([]string{"remove", "myapp"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(cfgDir, "config.toml"))
	if err != nil {
		t.Fatalf("config file missing: %v", err)
	}

	type configFile struct {
		Projects []struct {
			Name string `toml:"name"`
		} `toml:"projects"`
	}
	var cfg configFile
	if err := toml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project after remove, got %d", len(cfg.Projects))
	}
	if cfg.Projects[0].Name != "other" {
		t.Errorf("expected 'other', got %q", cfg.Projects[0].Name)
	}
}

func TestRemoveCommand_NotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cmd := project.Cmd
	cmd.SetArgs([]string{"remove", "nonexistent"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for nonexistent project, got nil")
	}
}
