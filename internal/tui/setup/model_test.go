package setup_test

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/ArielSurco/cli/internal/module"
	"github.com/ArielSurco/cli/internal/tui/setup"
)

func TestSetup_SaveWithCtrlS(t *testing.T) {
	allModules := module.Registry
	setupModel := setup.New(allModules, []string{})
	testModel := teatest.NewTestModel(t, setupModel, teatest.WithInitialTermSize(80, 24))

	teatest.WaitFor(t, testModel.Output(),
		func(output []byte) bool { return bytes.Contains(output, []byte("Setup: Active Modules")) },
		teatest.WithDuration(3*time.Second),
	)

	// Toggle first item with space
	testModel.Send(tea.KeyMsg{Type: tea.KeySpace})
	// Save with Ctrl+S
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlS})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(setup.Model)
	selectionResult := finalModel.Result()

	if !selectionResult.Saved {
		t.Error("expected Saved=true after Ctrl+S")
	}
	if len(selectionResult.ActiveModules) != 1 {
		t.Errorf("expected 1 active module, got %d", len(selectionResult.ActiveModules))
	}
	if selectionResult.ActiveModules[0] != allModules[0].Name {
		t.Errorf("expected active module %q, got %q", allModules[0].Name, selectionResult.ActiveModules[0])
	}
}

func TestSetup_CancelWithEsc(t *testing.T) {
	allModules := module.Registry
	setupModel := setup.New(allModules, []string{})
	testModel := teatest.NewTestModel(t, setupModel, teatest.WithInitialTermSize(80, 24))

	teatest.WaitFor(t, testModel.Output(),
		func(output []byte) bool { return bytes.Contains(output, []byte("Setup: Active Modules")) },
		teatest.WithDuration(3*time.Second),
	)

	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(setup.Model)
	selectionResult := finalModel.Result()

	if selectionResult.Saved {
		t.Error("expected Saved=false after Esc")
	}
}

func TestSetup_CancelWithCtrlC(t *testing.T) {
	allModules := module.Registry
	setupModel := setup.New(allModules, []string{})
	testModel := teatest.NewTestModel(t, setupModel, teatest.WithInitialTermSize(80, 24))

	teatest.WaitFor(t, testModel.Output(),
		func(output []byte) bool { return bytes.Contains(output, []byte("Setup: Active Modules")) },
		teatest.WithDuration(3*time.Second),
	)

	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(setup.Model)
	selectionResult := finalModel.Result()

	if selectionResult.Saved {
		t.Error("expected Saved=false after Ctrl+C")
	}
}

func TestSetup_PreChecked(t *testing.T) {
	allModules := module.Registry
	firstModuleName := allModules[0].Name
	setupModel := setup.New(allModules, []string{firstModuleName})
	testModel := teatest.NewTestModel(t, setupModel, teatest.WithInitialTermSize(80, 24))

	teatest.WaitFor(t, testModel.Output(),
		func(output []byte) bool { return bytes.Contains(output, []byte("Setup: Active Modules")) },
		teatest.WithDuration(3*time.Second),
	)

	// Save immediately without toggling — pre-checked module should still be active
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlS})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(setup.Model)
	selectionResult := finalModel.Result()

	if !selectionResult.Saved {
		t.Error("expected Saved=true")
	}
	if len(selectionResult.ActiveModules) != 1 {
		t.Errorf("expected 1 active module (pre-checked), got %d", len(selectionResult.ActiveModules))
	}
	if selectionResult.ActiveModules[0] != firstModuleName {
		t.Errorf("expected active module %q, got %q", firstModuleName, selectionResult.ActiveModules[0])
	}
}

func TestSetup_NavigateWithArrows(t *testing.T) {
	// Use two modules so we can navigate between them
	allModules := []module.Module{
		{Name: "alpha", Commands: []module.CommandDef{}},
		{Name: "beta", Commands: []module.CommandDef{}},
	}
	setupModel := setup.New(allModules, []string{})
	testModel := teatest.NewTestModel(t, setupModel, teatest.WithInitialTermSize(80, 24))

	teatest.WaitFor(t, testModel.Output(),
		func(output []byte) bool { return bytes.Contains(output, []byte("Setup: Active Modules")) },
		teatest.WithDuration(3*time.Second),
	)

	// Navigate down to the second item, then toggle it with space
	testModel.Send(tea.KeyMsg{Type: tea.KeyDown})
	testModel.Send(tea.KeyMsg{Type: tea.KeySpace})
	// Save with Ctrl+S
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlS})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(setup.Model)
	selectionResult := finalModel.Result()

	if !selectionResult.Saved {
		t.Error("expected Saved=true")
	}
	if len(selectionResult.ActiveModules) != 1 {
		t.Errorf("expected 1 active module (second item), got %d", len(selectionResult.ActiveModules))
	}
	if selectionResult.ActiveModules[0] != "beta" {
		t.Errorf("expected active module 'beta', got %q", selectionResult.ActiveModules[0])
	}
}
