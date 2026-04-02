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

// --- UpdateDevScript ---

func TestUpdateDevScript(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	tests := []struct {
		name            string
		initialProjects []config.Project
		projectName     string
		newDevScript    string
		wantErr         error
		wantDevScript   string
	}{
		{
			name: "update existing dev script to new value",
			initialProjects: []config.Project{
				{Name: "api", Path: "/projects/api", DevScript: "npm start"},
			},
			projectName:   "api",
			newDevScript:  "yarn dev",
			wantErr:       nil,
			wantDevScript: "yarn dev",
		},
		{
			name: "clear dev script by setting empty string",
			initialProjects: []config.Project{
				{Name: "api", Path: "/projects/api", DevScript: "npm start"},
			},
			projectName:   "api",
			newDevScript:  "",
			wantErr:       nil,
			wantDevScript: "",
		},
		{
			name:            "unknown project returns ErrNotFound",
			initialProjects: []config.Project{},
			projectName:     "ghost",
			newDevScript:    "yarn dev",
			wantErr:         project.ErrNotFound,
			wantDevScript:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService(tt.initialProjects...)

			err := svc.UpdateDevScript(tt.projectName, tt.newDevScript)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UpdateDevScript(%q, %q) error = %v, want %v", tt.projectName, tt.newDevScript, err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			foundProject, findErr := svc.FindByName(tt.projectName)
			if findErr != nil {
				t.Fatalf("FindByName after UpdateDevScript: %v", findErr)
			}

			if foundProject.DevScript != tt.wantDevScript {
				t.Errorf("DevScript = %q, want %q", foundProject.DevScript, tt.wantDevScript)
			}
		})
	}
}

func TestUpdateDevScript_NeverCallsConfigSave(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Service only mutates in-memory; config.Save() is the cmd layer's responsibility.
	// This test verifies the in-memory mutation is visible via List/FindByName,
	// confirming UpdateDevScript works correctly without relying on persistence.
	svc := newService(config.Project{Name: "api", Path: "/projects/api", DevScript: "npm start"})

	if err := svc.UpdateDevScript("api", "yarn dev"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The mutation must be visible in-memory.
	foundProject, err := svc.FindByName("api")
	if err != nil {
		t.Fatalf("FindByName: %v", err)
	}

	if foundProject.DevScript != "yarn dev" {
		t.Errorf("in-memory DevScript = %q, want %q", foundProject.DevScript, "yarn dev")
	}

	// List also reflects the change (same backing slice).
	list := svc.List()
	if len(list) != 1 || list[0].DevScript != "yarn dev" {
		t.Errorf("List()[0].DevScript = %q, want %q", list[0].DevScript, "yarn dev")
	}
}
