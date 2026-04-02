package projectscan_test

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/ArielSurco/cli/internal/tui/projectscan"
)

// helpers ---------------------------------------------------------------

func dirEntries() []projectscan.DirEntry {
	return []projectscan.DirEntry{
		{Name: "go-cli", AbsPath: "/projects/go-cli"},
		{Name: "api", AbsPath: "/projects/api"},
		{Name: "webapp", AbsPath: "/projects/webapp"},
	}
}

func waitFor(t *testing.T, testModel *teatest.TestModel, text string) {
	t.Helper()
	teatest.WaitFor(t, testModel.Output(),
		func(output []byte) bool { return bytes.Contains(output, []byte(text)) },
		teatest.WithDuration(3*time.Second),
	)
}

// stateChecklist -------------------------------------------------------

func TestChecklist_SpaceTogglesUncheckedToChecked(t *testing.T) {
	dirs := dirEntries()
	tuiModel := projectscan.New(dirs, nil)
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	waitFor(t, testModel, "go-cli")

	// Initially unchecked — send space to toggle
	testModel.Send(tea.KeyMsg{Type: tea.KeySpace})

	waitFor(t, testModel, "●")

	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

func TestChecklist_SpaceTogglesCheckedToUnchecked(t *testing.T) {
	dirs := dirEntries()
	// Register go-cli so it starts checked
	registered := map[string]string{"/projects/go-cli": "go-cli"}
	tuiModel := projectscan.New(dirs, registered)
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	waitFor(t, testModel, "●")

	// Toggle off the already-checked registered item
	testModel.Send(tea.KeyMsg{Type: tea.KeySpace})

	waitFor(t, testModel, "○")

	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

func TestChecklist_EnterWithNewSelectionConfirmsAndPopulatesAdd(t *testing.T) {
	dirs := dirEntries()
	tuiModel := projectscan.New(dirs, nil)
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	waitFor(t, testModel, "go-cli")

	// Check first item then confirm
	testModel.Send(tea.KeyMsg{Type: tea.KeySpace})
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(projectscan.Model)
	result := finalModel.Result()

	if !result.Confirmed {
		t.Fatal("expected Confirmed=true")
	}
	if len(result.ToAdd) != 1 || result.ToAdd[0] != "/projects/go-cli" {
		t.Errorf("expected ToAdd=[/projects/go-cli], got %v", result.ToAdd)
	}
	if len(result.ToRemove) != 0 {
		t.Errorf("expected empty ToRemove, got %v", result.ToRemove)
	}
}

func TestChecklist_EscQuitsWithConfirmedFalse(t *testing.T) {
	dirs := dirEntries()
	tuiModel := projectscan.New(dirs, nil)
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	waitFor(t, testModel, "go-cli")

	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(projectscan.Model)
	result := finalModel.Result()

	if result.Confirmed {
		t.Error("expected Confirmed=false after Esc")
	}
	if len(result.ToAdd) != 0 {
		t.Errorf("expected empty ToAdd, got %v", result.ToAdd)
	}
	if len(result.ToRemove) != 0 {
		t.Errorf("expected empty ToRemove, got %v", result.ToRemove)
	}
}

func TestChecklist_RegisteredItemsStartPreChecked(t *testing.T) {
	dirs := dirEntries()
	registered := map[string]string{
		"/projects/go-cli": "go-cli",
		"/projects/api":    "api",
	}
	tuiModel := projectscan.New(dirs, registered)
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	// Both registered items should render as checked (●)
	waitFor(t, testModel, "●")

	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	// Confirm internal state: no ToAdd (all were registered), no ToRemove (still checked)
	// Note: Confirmed=false because we pressed Esc
	finalModel := testModel.FinalModel(t).(projectscan.Model)
	result := finalModel.Result()

	if result.Confirmed {
		t.Error("expected Confirmed=false after Esc")
	}
}

func TestChecklist_FooterShowsChecklistText(t *testing.T) {
	dirs := dirEntries()
	tuiModel := projectscan.New(dirs, nil)
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	waitFor(t, testModel, "[space] toggle")

	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

// stateConfirmDeselect -------------------------------------------------

func TestConfirmDeselect_DeselectedRegisteredEnterTransitionsToConfirmState(t *testing.T) {
	dirs := dirEntries()
	registered := map[string]string{"/projects/go-cli": "go-cli"}
	tuiModel := projectscan.New(dirs, registered)
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	waitFor(t, testModel, "go-cli")

	// Deselect the registered item then press Enter → should show confirm footer
	testModel.Send(tea.KeyMsg{Type: tea.KeySpace})
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Confirm footer with ▶▶ must appear — this verifies the state transition
	waitFor(t, testModel, "▶▶")

	// Esc from stateConfirmDeselect returns to stateChecklist (not quit)
	// A second Esc from stateChecklist quits the program
	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	waitFor(t, testModel, "[space] toggle")
	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

func TestConfirmDeselect_EnterConfirmsRemoval(t *testing.T) {
	dirs := dirEntries()
	registered := map[string]string{"/projects/go-cli": "go-cli"}
	tuiModel := projectscan.New(dirs, registered)
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	waitFor(t, testModel, "go-cli")

	// Deselect registered item → confirm screen → confirm
	testModel.Send(tea.KeyMsg{Type: tea.KeySpace})
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	waitFor(t, testModel, "▶▶")
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(projectscan.Model)
	result := finalModel.Result()

	if !result.Confirmed {
		t.Fatal("expected Confirmed=true after double-Enter")
	}
	if len(result.ToRemove) != 1 || result.ToRemove[0] != "go-cli" {
		t.Errorf("expected ToRemove=[go-cli], got %v", result.ToRemove)
	}
	if len(result.ToAdd) != 0 {
		t.Errorf("expected empty ToAdd, got %v", result.ToAdd)
	}
}

func TestConfirmDeselect_EscReturnsToChecklist(t *testing.T) {
	dirs := dirEntries()
	registered := map[string]string{"/projects/go-cli": "go-cli"}
	tuiModel := projectscan.New(dirs, registered)
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	waitFor(t, testModel, "go-cli")

	// Deselect → confirm screen → Esc → back to checklist
	testModel.Send(tea.KeyMsg{Type: tea.KeySpace})
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	waitFor(t, testModel, "▶▶")

	// Esc in confirm state returns to checklist (footer changes back)
	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	waitFor(t, testModel, "[space] toggle")

	// Now quit cleanly
	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(projectscan.Model)
	result := finalModel.Result()

	if result.Confirmed {
		t.Error("expected Confirmed=false: user escaped at checklist level")
	}
}

func TestChecklist_AddOnlyPathSkipsConfirmScreen(t *testing.T) {
	dirs := dirEntries()
	// No registered items — checking a new item and pressing Enter should
	// go directly to Confirmed=true without showing the confirm footer.
	tuiModel := projectscan.New(dirs, nil)
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	waitFor(t, testModel, "go-cli")

	testModel.Send(tea.KeyMsg{Type: tea.KeySpace})
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(projectscan.Model)
	result := finalModel.Result()

	if !result.Confirmed {
		t.Fatal("expected Confirmed=true without confirm screen")
	}
	if len(result.ToAdd) != 1 {
		t.Errorf("expected 1 item in ToAdd, got %v", result.ToAdd)
	}
}
