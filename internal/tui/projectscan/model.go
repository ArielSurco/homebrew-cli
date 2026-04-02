package projectscan

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DirEntry represents a scanned directory.
type DirEntry struct {
	Name    string
	AbsPath string
}

// scanItem is an internal item in the checklist.
type scanItem struct {
	entry          DirEntry
	isRegistered   bool
	checked        bool
	registeredName string // project name from the registeredPaths map (empty if not registered)
}

type scanState int

const (
	stateChecklist       scanState = iota
	stateConfirmDeselect           // shown when Enter would remove registered projects
)

// Result is returned after the TUI exits.
type Result struct {
	ToAdd     []string // absolute paths of newly selected dirs
	ToRemove  []string // project names of deselected registered projects
	Confirmed bool
}

// Model is the Bubbletea model for directory scan selection.
type Model struct {
	items       []scanItem
	cursorIndex int
	state       scanState
	result      Result
}

// New creates the model.
// registeredPaths maps absolute path -> project name for already-registered dirs.
// Items whose AbsPath is a key in registeredPaths start pre-checked with isRegistered=true.
func New(dirs []DirEntry, registeredPaths map[string]string) Model {
	items := make([]scanItem, len(dirs))
	for index, dirEntry := range dirs {
		projectName, alreadyRegistered := registeredPaths[dirEntry.AbsPath]
		items[index] = scanItem{
			entry:          dirEntry,
			isRegistered:   alreadyRegistered,
			checked:        alreadyRegistered,
			registeredName: projectName,
		}
	}
	return Model{
		items:       items,
		cursorIndex: 0,
		state:       stateChecklist,
	}
}

// Init implements tea.Model.
func (model Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return model, nil
	}

	switch model.state {
	case stateChecklist:
		return model.updateChecklist(keyMsg)
	case stateConfirmDeselect:
		return model.updateConfirmDeselect(keyMsg)
	}

	return model, nil
}

func (model Model) updateChecklist(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyMsg.Type {
	case tea.KeyUp:
		if model.cursorIndex > 0 {
			model.cursorIndex--
		}
	case tea.KeyDown:
		if model.cursorIndex < len(model.items)-1 {
			model.cursorIndex++
		}
	case tea.KeySpace:
		if len(model.items) > 0 {
			model.items[model.cursorIndex].checked = !model.items[model.cursorIndex].checked
		}
	case tea.KeyEnter:
		if model.hasDeselectedRegistered() {
			model.state = stateConfirmDeselect
			return model, nil
		}
		model.result = model.buildResult()
		return model, tea.Quit
	case tea.KeyEsc:
		model.result = Result{Confirmed: false}
		return model, tea.Quit
	case tea.KeyRunes:
		switch string(keyMsg.Runes) {
		case "k":
			if model.cursorIndex > 0 {
				model.cursorIndex--
			}
		case "j":
			if model.cursorIndex < len(model.items)-1 {
				model.cursorIndex++
			}
		}
	}

	return model, nil
}

func (model Model) updateConfirmDeselect(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyMsg.Type {
	case tea.KeyEnter:
		model.result = model.buildResult()
		return model, tea.Quit
	case tea.KeyEsc:
		// Return to checklist — re-check any deselected registered items
		for index := range model.items {
			if model.items[index].isRegistered && !model.items[index].checked {
				model.items[index].checked = true
			}
		}
		model.state = stateChecklist
	}

	return model, nil
}

// hasDeselectedRegistered reports whether any registered item is now unchecked.
func (model Model) hasDeselectedRegistered() bool {
	for _, item := range model.items {
		if item.isRegistered && !item.checked {
			return true
		}
	}
	return false
}

// buildResult constructs the Result from current checked state.
func (model Model) buildResult() Result {
	toAdd := make([]string, 0)
	toRemove := make([]string, 0)

	for _, item := range model.items {
		if item.checked && !item.isRegistered {
			toAdd = append(toAdd, item.entry.AbsPath)
		}
		if item.isRegistered && !item.checked {
			toRemove = append(toRemove, item.registeredName)
		}
	}

	return Result{
		ToAdd:     toAdd,
		ToRemove:  toRemove,
		Confirmed: true,
	}
}

// View implements tea.Model.
func (model Model) View() string {
	faintStyle := lipgloss.NewStyle().Faint(true)
	coralStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("204"))

	output := ""

	for index, item := range model.items {
		cursorPrefix := "  "
		if index == model.cursorIndex {
			cursorPrefix = "> "
		}

		checkMark := "○"
		if item.checked {
			checkMark = "●"
		}

		output += fmt.Sprintf("%s%s %s\n", cursorPrefix, checkMark, item.entry.Name)
	}

	switch model.state {
	case stateChecklist:
		output += "\n" + faintStyle.Render("[space] toggle  [enter] apply  [esc] cancel")
	case stateConfirmDeselect:
		output += "\n" + coralStyle.Render("▶▶ removing registered projects — [enter] confirm  [esc] cancel")
	}

	return output
}

// Result returns the selection outcome after the program exits.
func (model Model) Result() Result {
	return model.result
}
