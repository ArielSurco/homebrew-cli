package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	cmdproject "github.com/ArielSurco/cli/cmd/project"
	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/tui/projectscan"
)

// --- scanSubdirectories tests ---

func TestScanSubdirectories_ReturnsDirEntries(t *testing.T) {
	rootDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(rootDir, "alpha"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "beta"), 0o755); err != nil {
		t.Fatal(err)
	}

	entries, err := cmdproject.ScanSubdirectories(rootDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	names := map[string]bool{}
	for _, entry := range entries {
		names[entry.Name] = true
		expectedAbsPath := filepath.Join(rootDir, entry.Name)
		if entry.AbsPath != expectedAbsPath {
			t.Errorf("entry %q: expected AbsPath %q, got %q", entry.Name, expectedAbsPath, entry.AbsPath)
		}
	}

	if !names["alpha"] || !names["beta"] {
		t.Errorf("expected entries 'alpha' and 'beta', got %v", names)
	}
}

func TestScanSubdirectories_SkipsHiddenDirs(t *testing.T) {
	rootDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(rootDir, ".hidden"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "visible"), 0o755); err != nil {
		t.Fatal(err)
	}

	entries, err := cmdproject.ScanSubdirectories(rootDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d: %v", len(entries), entries)
	}
	if entries[0].Name != "visible" {
		t.Errorf("expected entry 'visible', got %q", entries[0].Name)
	}
}

func TestScanSubdirectories_SkipsRegularFiles(t *testing.T) {
	rootDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(rootDir, "file.txt"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	entries, err := cmdproject.ScanSubdirectories(rootDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "subdir" {
		t.Errorf("expected entry 'subdir', got %q", entries[0].Name)
	}
}

func TestScanSubdirectories_SkipsSymlinks(t *testing.T) {
	rootDir := t.TempDir()
	targetDir := t.TempDir()

	symlinkPath := filepath.Join(rootDir, "linked")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "realdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	entries, err := cmdproject.ScanSubdirectories(rootDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (symlink skipped), got %d: %v", len(entries), entries)
	}
	if entries[0].Name != "realdir" {
		t.Errorf("expected entry 'realdir', got %q", entries[0].Name)
	}
}

func TestScanSubdirectories_ErrorWhenRootDoesNotExist(t *testing.T) {
	_, err := cmdproject.ScanSubdirectories("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for nonexistent rootPath, got nil")
	}
}

func TestScanSubdirectories_EmptySliceWhenNoEligibleSubdirs(t *testing.T) {
	rootDir := t.TempDir()

	entries, err := cmdproject.ScanSubdirectories(rootDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(entries))
	}
}

// --- applyWatchResult tests ---

func TestApplyWatchResult_NotConfirmed_NoSvcCalls(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "existing", Path: "/some/path"},
		},
	}

	result := projectscan.Result{
		Confirmed: false,
		ToAdd:     []string{"/new/path"},
		ToRemove:  []string{"existing"},
	}

	if err := cmdproject.ApplyWatchResult(result, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Config must be unchanged — no projects were added or removed
	if len(cfg.Projects) != 1 {
		t.Errorf("expected 1 project unchanged, got %d", len(cfg.Projects))
	}
	if cfg.Projects[0].Name != "existing" {
		t.Errorf("expected project 'existing', got %q", cfg.Projects[0].Name)
	}
}

func TestApplyWatchResult_ToAdd_CallsServiceAdd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{Projects: []config.Project{}}

	newPath := t.TempDir()
	dirName := filepath.Base(newPath)

	result := projectscan.Result{
		Confirmed: true,
		ToAdd:     []string{newPath},
		ToRemove:  []string{},
	}

	if err := cmdproject.ApplyWatchResult(result, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project added, got %d", len(cfg.Projects))
	}
	if cfg.Projects[0].Name != dirName {
		t.Errorf("expected project name %q, got %q", dirName, cfg.Projects[0].Name)
	}
	if cfg.Projects[0].Path != newPath {
		t.Errorf("expected project path %q, got %q", newPath, cfg.Projects[0].Path)
	}
}

func TestApplyWatchResult_ToRemove_CallsServiceRemove(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "myproject", Path: "/some/path"},
		},
	}

	result := projectscan.Result{
		Confirmed: true,
		ToAdd:     []string{},
		ToRemove:  []string{"myproject"},
	}

	if err := cmdproject.ApplyWatchResult(result, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Projects) != 0 {
		t.Errorf("expected 0 projects after removal, got %d", len(cfg.Projects))
	}
}

func TestApplyWatchResult_BothToAddAndToRemove(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: "old-project", Path: "/old/path"},
		},
	}

	newPath := t.TempDir()
	dirName := filepath.Base(newPath)

	result := projectscan.Result{
		Confirmed: true,
		ToAdd:     []string{newPath},
		ToRemove:  []string{"old-project"},
	}

	if err := cmdproject.ApplyWatchResult(result, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project after add+remove, got %d", len(cfg.Projects))
	}
	if cfg.Projects[0].Name != dirName {
		t.Errorf("expected new project %q, got %q", dirName, cfg.Projects[0].Name)
	}
}

func TestApplyWatchResult_ToAdd_DuplicateName_IsNonFatal(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	existingPath := t.TempDir()
	existingName := filepath.Base(existingPath)

	cfg := &config.Config{
		Projects: []config.Project{
			{Name: existingName, Path: existingPath},
		},
	}

	result := projectscan.Result{
		Confirmed: true,
		ToAdd:     []string{existingPath},
		ToRemove:  []string{},
	}

	err := cmdproject.ApplyWatchResult(result, cfg)
	if err != nil {
		t.Fatalf("ErrDuplicateName should be non-fatal, got: %v", err)
	}

	// The original project should still be there, not doubled
	if len(cfg.Projects) != 1 {
		t.Errorf("expected 1 project (no duplicate), got %d", len(cfg.Projects))
	}
}

func TestApplyWatchResult_ToRemove_NotFound_IsNonFatal(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg := &config.Config{Projects: []config.Project{}}

	result := projectscan.Result{
		Confirmed: true,
		ToAdd:     []string{},
		ToRemove:  []string{"ghost-project"},
	}

	err := cmdproject.ApplyWatchResult(result, cfg)
	if err != nil {
		t.Fatalf("ErrNotFound should be non-fatal, got: %v", err)
	}
}

func TestApplyWatchResult_SavesConfig(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	cfg := &config.Config{Projects: []config.Project{}}

	newPath := t.TempDir()

	result := projectscan.Result{
		Confirmed: true,
		ToAdd:     []string{newPath},
		ToRemove:  []string{},
	}

	if err := cmdproject.ApplyWatchResult(result, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify config was saved by loading it back
	savedCfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if len(savedCfg.Projects) != 1 {
		t.Errorf("expected 1 project in saved config, got %d", len(savedCfg.Projects))
	}
}

// Verify that the errors package is used (compilation check).
var _ = errors.New
