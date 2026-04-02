package module_test

import (
	"testing"

	"github.com/ArielSurco/cli/internal/module"
)

func TestRegistry_HasAtLeastOneModule(t *testing.T) {
	if len(module.Registry) == 0 {
		t.Error("Registry must have at least one module")
	}
}

func TestRegistry_GoProjectModuleExists(t *testing.T) {
	goProject, found := module.FindModule("go-project")
	if !found {
		t.Fatal("expected to find module 'go-project'")
	}
	if len(goProject.Commands) != 5 {
		t.Errorf("expected 5 commands in go-project, got %d", len(goProject.Commands))
	}
}

func TestRegistry_NoDuplicateAliases(t *testing.T) {
	seen := make(map[string]bool)
	for _, registryModule := range module.Registry {
		for _, commandDef := range registryModule.Commands {
			if seen[commandDef.Alias] {
				t.Errorf("duplicate alias %q found in registry", commandDef.Alias)
			}
			seen[commandDef.Alias] = true
		}
	}
}

func TestRegistry_CommandFlags(t *testing.T) {
	tests := []struct {
		name            string
		alias           string
		needsEval       bool
		hasCompletions  bool
	}{
		{name: "gp: NeedsEval=false HasCompletions=true", alias: "gp", needsEval: false, hasCompletions: true},
		{name: "gpd: NeedsEval=true HasCompletions=true", alias: "gpd", needsEval: true, hasCompletions: true},
		{name: "gpa: NeedsEval=false HasCompletions=false", alias: "gpa", needsEval: false, hasCompletions: false},
		{name: "gpr: NeedsEval=false HasCompletions=true", alias: "gpr", needsEval: false, hasCompletions: true},
	}

	goProject, found := module.FindModule("go-project")
	if !found {
		t.Fatal("go-project module not found")
	}

	commandsByAlias := make(map[string]module.CommandDef)
	for _, commandDef := range goProject.Commands {
		commandsByAlias[commandDef.Alias] = commandDef
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			commandDef, exists := commandsByAlias[testCase.alias]
			if !exists {
				t.Fatalf("alias %q not found in go-project", testCase.alias)
			}
			if commandDef.NeedsEval != testCase.needsEval {
				t.Errorf("alias %q: NeedsEval = %v, want %v", testCase.alias, commandDef.NeedsEval, testCase.needsEval)
			}
			if commandDef.HasCompletions != testCase.hasCompletions {
				t.Errorf("alias %q: HasCompletions = %v, want %v", testCase.alias, commandDef.HasCompletions, testCase.hasCompletions)
			}
		})
	}
}

func TestFindModule_Found(t *testing.T) {
	result, found := module.FindModule("go-project")
	if !found {
		t.Fatal("FindModule(\"go-project\") returned false, expected true")
	}
	if result == nil {
		t.Fatal("FindModule(\"go-project\") returned nil module")
	}
	if result.Name != "go-project" {
		t.Errorf("got module name %q, want %q", result.Name, "go-project")
	}
}

func TestFindModule_NotFound(t *testing.T) {
	_, found := module.FindModule("nonexistent")
	if found {
		t.Error("FindModule(\"nonexistent\") returned true, expected false")
	}
}

func TestActiveModules_ReturnsMatchingModules(t *testing.T) {
	result := module.ActiveModules([]string{"go-project"})
	if len(result) != 1 {
		t.Errorf("ActiveModules([\"go-project\"]) returned %d modules, want 1", len(result))
	}
	if result[0].Name != "go-project" {
		t.Errorf("got module name %q, want %q", result[0].Name, "go-project")
	}
}

func TestActiveModules_EmptyInput(t *testing.T) {
	result := module.ActiveModules([]string{})
	if len(result) != 0 {
		t.Errorf("ActiveModules([]) returned %d modules, want 0", len(result))
	}
}

func TestActiveModules_UnknownName(t *testing.T) {
	result := module.ActiveModules([]string{"unknown"})
	if len(result) != 0 {
		t.Errorf("ActiveModules([\"unknown\"]) returned %d modules, want 0", len(result))
	}
}
