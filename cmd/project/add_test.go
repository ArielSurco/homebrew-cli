package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ArielSurco/cli/cmd/project"
	"github.com/pelletier/go-toml/v2"
)

func TestAddCommand_Success(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cmd := project.Cmd
	cmd.SetArgs([]string{"add", "myapp", "/absolute/path", "--dev-script", "npm run dev"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify config file was written
	cfgPath := filepath.Join(dir, "arielsurco-cli", "config.toml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("config file not written: %v", err)
	}

	type configFile struct {
		Projects []struct {
			Name      string `toml:"name"`
			Path      string `toml:"path"`
			DevScript string `toml:"dev_script"`
		} `toml:"projects"`
	}
	var cfg configFile
	if err := toml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(cfg.Projects))
	}
	p := cfg.Projects[0]
	if p.Name != "myapp" || p.Path != "/absolute/path" || p.DevScript != "npm run dev" {
		t.Errorf("unexpected project: %+v", p)
	}
}

func TestAddCommand_DuplicateName(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Pre-populate config
	cfgDir := filepath.Join(dir, "arielsurco-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := `[[projects]]
name = "myapp"
path = "/existing/path"
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := project.Cmd
	cmd.SetArgs([]string{"add", "myapp", "/other/path"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAddCommand_RelativePath_ResolvesToAbsolute(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cmd := project.Cmd
	cmd.SetArgs([]string{"add", "myapp", "."})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected relative path to be resolved, got error: %v", err)
	}
}
