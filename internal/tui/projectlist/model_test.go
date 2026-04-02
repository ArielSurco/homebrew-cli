package projectlist_test

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/ArielSurco/cli/internal/tui/projectlist"
)

func TestProjectList_SelectWithEnter(t *testing.T) {
	projects := []config.Project{
		{Name: "go-cli", Path: "/projects/go-cli"},
		{Name: "api", Path: "/projects/api"},
	}

	tuiModel := projectlist.New(projects, "")
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	// Wait for initial render showing the first project
	teatest.WaitFor(t, testModel.Output(),
		func(output []byte) bool { return bytes.Contains(output, []byte("go-cli")) },
		teatest.WithDuration(3*time.Second),
	)

	// Press enter to select first item
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(projectlist.Model)
	selectionResult := finalModel.Result()

	if selectionResult.Cancelled {
		t.Error("expected not cancelled")
	}
	if selectionResult.Project.Name != "go-cli" {
		t.Errorf("expected project name go-cli, got %q", selectionResult.Project.Name)
	}
	if selectionResult.Project.Path != "/projects/go-cli" {
		t.Errorf("expected project path /projects/go-cli, got %q", selectionResult.Project.Path)
	}
}

func TestProjectList_CancelWithEsc(t *testing.T) {
	projects := []config.Project{
		{Name: "go-cli", Path: "/projects/go-cli"},
		{Name: "api", Path: "/projects/api"},
	}

	tuiModel := projectlist.New(projects, "")
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	teatest.WaitFor(t, testModel.Output(),
		func(output []byte) bool { return bytes.Contains(output, []byte("go-cli")) },
		teatest.WithDuration(3*time.Second),
	)

	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(projectlist.Model)
	selectionResult := finalModel.Result()

	if !selectionResult.Cancelled {
		t.Error("expected cancelled=true after Esc")
	}
}

func TestProjectList_CancelWithCtrlC(t *testing.T) {
	projects := []config.Project{
		{Name: "go-cli", Path: "/projects/go-cli"},
	}

	tuiModel := projectlist.New(projects, "")
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	teatest.WaitFor(t, testModel.Output(),
		func(output []byte) bool { return bytes.Contains(output, []byte("go-cli")) },
		teatest.WithDuration(3*time.Second),
	)

	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(projectlist.Model)
	selectionResult := finalModel.Result()

	if !selectionResult.Cancelled {
		t.Error("expected cancelled=true after Ctrl+C")
	}
}

func TestProjectList_PreFilterSeed(t *testing.T) {
	projects := []config.Project{
		{Name: "go-cli", Path: "/projects/go-cli"},
		{Name: "api", Path: "/projects/api"},
		{Name: "webapp", Path: "/projects/webapp"},
	}

	// Pre-filter to "api" — only "api" should be visible
	tuiModel := projectlist.New(projects, "api")
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	teatest.WaitFor(t, testModel.Output(),
		func(output []byte) bool { return bytes.Contains(output, []byte("api")) },
		teatest.WithDuration(3*time.Second),
	)

	// Select the filtered item
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(projectlist.Model)
	selectionResult := finalModel.Result()

	if selectionResult.Cancelled {
		t.Error("expected not cancelled")
	}
	if selectionResult.Project.Name != "api" {
		t.Errorf("expected project name api (from prefilter), got %q", selectionResult.Project.Name)
	}
}

func TestProjectList_EmptyProjects(t *testing.T) {
	tuiModel := projectlist.New([]config.Project{}, "")
	testModel := teatest.NewTestModel(t, tuiModel, teatest.WithInitialTermSize(80, 24))

	// With no items, send Esc to quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	testModel.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	finalModel := testModel.FinalModel(t).(projectlist.Model)
	selectionResult := finalModel.Result()

	if !selectionResult.Cancelled {
		t.Error("expected cancelled=true for empty project list")
	}
}
