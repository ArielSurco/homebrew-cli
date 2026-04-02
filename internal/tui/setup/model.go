package setup

import (
	"fmt"

	"github.com/arielsurco/go-cli/internal/module"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Result holds the outcome of the setup TUI.
type Result struct {
	ActiveModules []string // module names the user confirmed active
	Saved         bool     // true if user pressed Ctrl+S, false if cancelled
}

// Model is the Bubbletea model for module selection.
type Model struct {
	modules       []module.Module
	checkedStates map[string]bool // module name → checked
	cursorIndex   int
	saved         bool
}

// New creates a setup model.
// allModules: full registry to display
// currentActive: module names that should start checked
func New(allModules []module.Module, currentActive []string) Model {
	checkedStates := make(map[string]bool, len(allModules))
	for _, activeName := range currentActive {
		checkedStates[activeName] = true
	}
	return Model{
		modules:       allModules,
		checkedStates: checkedStates,
		cursorIndex:   0,
		saved:         false,
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

	switch keyMsg.Type {
	case tea.KeyUp:
		if model.cursorIndex > 0 {
			model.cursorIndex--
		}
	case tea.KeyDown:
		if model.cursorIndex < len(model.modules)-1 {
			model.cursorIndex++
		}
	case tea.KeySpace, tea.KeyEnter:
		if len(model.modules) > 0 {
			currentName := model.modules[model.cursorIndex].Name
			model.checkedStates[currentName] = !model.checkedStates[currentName]
		}
	case tea.KeyCtrlS:
		model.saved = true
		return model, tea.Quit
	case tea.KeyEsc, tea.KeyCtrlC:
		return model, tea.Quit
	case tea.KeyRunes:
		switch string(keyMsg.Runes) {
		case "k":
			if model.cursorIndex > 0 {
				model.cursorIndex--
			}
		case "j":
			if model.cursorIndex < len(model.modules)-1 {
				model.cursorIndex++
			}
		}
	}

	return model, nil
}

// View implements tea.Model.
func (model Model) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true)
	checkedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	footerStyle := lipgloss.NewStyle().Faint(true)

	output := titleStyle.Render("Setup: Active Modules") + "\n\n"

	for index, moduleDef := range model.modules {
		cursor := "  "
		if index == model.cursorIndex {
			cursor = "> "
		}

		checkbox := "[ ]"
		moduleName := moduleDef.Name
		if model.checkedStates[moduleDef.Name] {
			checkbox = checkedStyle.Render("[x]")
			moduleName = checkedStyle.Render(moduleDef.Name)
		}

		output += fmt.Sprintf("%s%s %s\n", cursor, checkbox, moduleName)
	}

	output += footerStyle.Render("\nctrl+s: save   esc: cancel")
	return output
}

// Result returns the selection outcome after the program exits.
func (model Model) Result() Result {
	activeModuleNames := make([]string, 0)
	for _, moduleDef := range model.modules {
		if model.checkedStates[moduleDef.Name] {
			activeModuleNames = append(activeModuleNames, moduleDef.Name)
		}
	}
	return Result{ActiveModules: activeModuleNames, Saved: model.saved}
}
