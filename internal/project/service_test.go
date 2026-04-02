package project_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/project"
)

func newService(projects ...config.Project) *project.Service {
	cfg := &config.Config{Projects: projects}
	return project.NewService(cfg)
}

// --- Add ---

func TestAdd_Success(t *testing.T) {
	svc := newService()
	if err := svc.Add("myapp", "/absolute/path", "npm start"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	list := svc.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 project, got %d", len(list))
	}
	if list[0].Name != "myapp" || list[0].Path != "/absolute/path" || list[0].DevScript != "npm start" {
		t.Errorf("unexpected project: %+v", list[0])
	}
}

func TestAdd_Duplicate(t *testing.T) {
	svc := newService(config.Project{Name: "myapp", Path: "/path"})
	err := svc.Add("myapp", "/other/path", "")
	if !errors.Is(err, project.ErrDuplicateName) {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

func TestAdd_RelativePath_ResolvesToAbsolute(t *testing.T) {
	svc := newService()
	if err := svc.Add("myapp", "relative/path", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	addedProject := svc.List()[0]
	if !filepath.IsAbs(addedProject.Path) {
		t.Errorf("expected absolute path, got %q", addedProject.Path)
	}
}

func TestAdd_DotPath_ResolvesToCWD(t *testing.T) {
	svc := newService()
	if err := svc.Add("myapp", ".", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	addedProject := svc.List()[0]
	if !filepath.IsAbs(addedProject.Path) {
		t.Errorf("expected absolute path from '.', got %q", addedProject.Path)
	}
}

// --- Remove ---

func TestRemove_Success(t *testing.T) {
	svc := newService(
		config.Project{Name: "keep", Path: "/keep"},
		config.Project{Name: "remove-me", Path: "/remove"},
	)
	if err := svc.Remove("remove-me"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	list := svc.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 project after remove, got %d", len(list))
	}
	if list[0].Name != "keep" {
		t.Errorf("wrong project remaining: %+v", list[0])
	}
}

func TestRemove_NotFound(t *testing.T) {
	svc := newService()
	err := svc.Remove("nonexistent")
	if !errors.Is(err, project.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- FindByName ---

func TestFindByName_Found(t *testing.T) {
	svc := newService(config.Project{Name: "myapp", Path: "/path", DevScript: "make"})
	foundProject, err := svc.FindByName("myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if foundProject.Name != "myapp" {
		t.Errorf("got %q, want %q", foundProject.Name, "myapp")
	}
}

func TestFindByName_NotFound(t *testing.T) {
	svc := newService()
	_, err := svc.FindByName("ghost")
	if !errors.Is(err, project.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- List ---

func TestList_ReturnsAll(t *testing.T) {
	projects := []config.Project{
		{Name: "a", Path: "/a"},
		{Name: "b", Path: "/b"},
		{Name: "c", Path: "/c"},
	}
	svc := newService(projects...)
	list := svc.List()
	if len(list) != 3 {
		t.Fatalf("expected 3, got %d", len(list))
	}
}

// --- DevCommand ---

func TestDevCommand_NormalPath(t *testing.T) {
	svc := newService(config.Project{Name: "app", Path: "/home/user/app", DevScript: "npm run dev"})
	shellCommand, err := svc.DevCommand("app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedCommand := "cd '/home/user/app' && npm run dev"
	if shellCommand != expectedCommand {
		t.Errorf("got %q, want %q", shellCommand, expectedCommand)
	}
}

func TestDevCommand_PathWithSpaces(t *testing.T) {
	svc := newService(config.Project{Name: "app", Path: "/home/my user/my app", DevScript: "make dev"})
	shellCommand, err := svc.DevCommand("app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedCommand := "cd '/home/my user/my app' && make dev"
	if shellCommand != expectedCommand {
		t.Errorf("got %q, want %q", shellCommand, expectedCommand)
	}
}

func TestDevCommand_PathWithSingleQuote(t *testing.T) {
	svc := newService(config.Project{Name: "app", Path: "/home/it's/here", DevScript: "yarn dev"})
	shellCommand, err := svc.DevCommand("app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedCommand := "cd '/home/it'\\''s/here' && yarn dev"
	if shellCommand != expectedCommand {
		t.Errorf("got %q, want %q", shellCommand, expectedCommand)
	}
}

func TestDevCommand_NoDevScript(t *testing.T) {
	svc := newService(config.Project{Name: "app", Path: "/path"})
	_, err := svc.DevCommand("app")
	if !errors.Is(err, project.ErrNoDevScript) {
		t.Errorf("expected ErrNoDevScript, got %v", err)
	}
}
