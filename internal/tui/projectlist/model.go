package projectlist

import (
	"fmt"

	"github.com/ArielSurco/cli/internal/config"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// footerMode represents the current interaction mode of the project list.
type footerMode int

const (
	modeNavigate      footerMode = iota
	modeConfirmDelete footerMode = iota
	modeConfirmEdit   footerMode = iota
)

// ActionKind describes the action the user selected.
type ActionKind int

const (
	// ActionNone means the user cancelled without selecting an action.
	ActionNone ActionKind = iota
	// ActionNavigate means the user wants to navigate to the project.
	ActionNavigate ActionKind = iota
	// ActionDelete means the user confirmed project deletion.
	ActionDelete ActionKind = iota
	// ActionEditDev means the user wants to edit the dev script for the project.
	ActionEditDev ActionKind = iota
)

// Result holds the outcome of the project selection TUI.
type Result struct {
	Project   config.Project
	Action    ActionKind
	Cancelled bool
}

// Model is the Bubbletea model for project selection.
type Model struct {
	list        list.Model
	mode        footerMode
	deleteMode  bool
	pendingItem *projectItem
	result      Result
}

type projectItem struct {
	project config.Project
}

func (item projectItem) Title() string {
	if item.project.DevScript != "" {
		return "● " + item.project.Name
	}
	return "○ " + item.project.Name
}

func (item projectItem) Description() string { return item.project.Path }
func (item projectItem) FilterValue() string  { return item.project.Name }

var (
	footerNavigateStyle = lipgloss.NewStyle().Faint(true)
	footerDeleteStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	footerEditStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)

const (
	// listWidth is the fixed width of the project list.
	listWidth = 80
	// listHeight is fixed to always fit up to 12 items without pagination.
	listHeight = 20
)

// New creates a projectlist model. If preFilter is non-empty, it is applied
// as the initial filter text so the list starts narrowed.
func New(projects []config.Project, preFilter string) Model {
	return newModel(projects, preFilter, false)
}

// NewForDelete creates a projectlist model in delete mode.
// Enter on a project goes directly to the delete confirmation footer
// instead of navigating, making it clear the intent is removal.
func NewForDelete(projects []config.Project, preFilter string) Model {
	return newModel(projects, preFilter, true)
}

func newModel(projects []config.Project, preFilter string, deleteMode bool) Model {
	items := make([]list.Item, len(projects))
	for index, existingProject := range projects {
		items[index] = &projectItem{project: existingProject}
	}

	compactDelegate := list.NewDefaultDelegate()
	compactDelegate.ShowDescription = false
	compactDelegate.SetSpacing(0)

	title := "Select a project"
	if deleteMode {
		title = "Remove a project"
	}

	listModel := list.New(items, compactDelegate, listWidth, listHeight)
	listModel.Title = title
	listModel.SetShowStatusBar(false)

	if preFilter != "" {
		listModel.SetFilteringEnabled(true)
		listModel.SetFilterText(preFilter)
	}

	return Model{list: listModel, deleteMode: deleteMode}
}

// Init implements tea.Model.
func (model Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// When list is in filter mode, forward all keys to the list — no custom handling.
		if model.list.FilterState() == list.Filtering {
			updatedList, listCmd := model.list.Update(msg)
			model.list = updatedList
			return model, listCmd
		}

		switch model.mode {
		case modeNavigate:
			switch keyMsg.Type {
			case tea.KeyEnter:
				selectedListItem := model.list.SelectedItem()
				if selectedListItem != nil {
					selected := selectedListItem.(*projectItem)
					if model.deleteMode {
						model.pendingItem = selected
						model.mode = modeConfirmDelete
						return model, nil
					}
					model.result = Result{
						Project: selected.project,
						Action:  ActionNavigate,
					}
					return model, tea.Quit
				}
			case tea.KeyCtrlC, tea.KeyEsc:
				model.result = Result{Action: ActionNone}
				return model, tea.Quit
			case tea.KeyRunes:
				switch string(keyMsg.Runes) {
				case "d":
					selectedListItem := model.list.SelectedItem()
					if selectedListItem != nil {
						selected := selectedListItem.(*projectItem)
						model.pendingItem = selected
						model.mode = modeConfirmDelete
						return model, nil
					}
				case "e":
					selectedListItem := model.list.SelectedItem()
					if selectedListItem != nil {
						selected := selectedListItem.(*projectItem)
						model.pendingItem = selected
						model.mode = modeConfirmEdit
						return model, nil
					}
				}
			}

		case modeConfirmDelete:
			// Do not forward key messages to model.list.Update() in confirm modes.
			switch keyMsg.Type {
			case tea.KeyEsc:
				model.mode = modeNavigate
				model.pendingItem = nil
				return model, nil
			case tea.KeyRunes:
				if string(keyMsg.Runes) == "y" && model.pendingItem != nil {
					model.result = Result{
						Project: model.pendingItem.project,
						Action:  ActionDelete,
					}
					return model, tea.Quit
				}
			}
			return model, nil

		case modeConfirmEdit:
			// Do not forward key messages to model.list.Update() in confirm modes.
			switch keyMsg.Type {
			case tea.KeyEsc:
				model.mode = modeNavigate
				model.pendingItem = nil
				return model, nil
			case tea.KeyEnter:
				if model.pendingItem != nil {
					model.result = Result{
						Project: model.pendingItem.project,
						Action:  ActionEditDev,
					}
					return model, tea.Quit
				}
			}
			return model, nil
		}
	}

	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		// Only update width — height stays content-driven, not terminal-driven.
		model.list.SetSize(sizeMsg.Width, model.list.Height())
	}

	updatedList, listCmd := model.list.Update(msg)
	model.list = updatedList
	return model, listCmd
}

// renderFooter builds the footer string based on current mode.
func (model Model) renderFooter() string {
	switch model.mode {
	case modeConfirmDelete:
		projectName := ""
		if model.pendingItem != nil {
			projectName = model.pendingItem.project.Name
		}
		return footerDeleteStyle.Render(
			fmt.Sprintf("▶▶ [d] remove '%s' — [y] confirm  [esc] cancel", projectName),
		)
	case modeConfirmEdit:
		projectName := ""
		if model.pendingItem != nil {
			projectName = model.pendingItem.project.Name
		}
		return footerEditStyle.Render(
			fmt.Sprintf("▶▶ [e] edit dev script '%s' — [enter] open editor  [esc] cancel", projectName),
		)
	default:
		if model.deleteMode {
			return footerDeleteStyle.Render(
				"select project to remove — [enter] confirm  [esc] cancel",
			)
		}
		return footerNavigateStyle.Render(
			"[d] remove  [e] edit dev script  [enter] navigate  [esc] cancel",
		)
	}
}

// View implements tea.Model.
func (model Model) View() string {
	return model.list.View() + "\n" + model.renderFooter()
}

// Result returns the selection outcome after the program exits.
func (model Model) Result() Result {
	// Backwards compatibility: set Cancelled when ActionNone or no project selected.
	result := model.result
	if result.Action == ActionNone {
		result.Cancelled = true
	}
	return result
}
