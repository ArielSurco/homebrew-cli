package projectlist

import (
	"github.com/ArielSurco/cli/internal/config"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Result holds the outcome of the project selection TUI.
type Result struct {
	Project   config.Project
	Cancelled bool
}

// Model is the Bubbletea model for project selection.
type Model struct {
	list          list.Model
	selectedItem  *projectItem // nil until user confirms
	cancelled     bool
}

type projectItem struct {
	project config.Project
}

func (item projectItem) Title() string       { return item.project.Name }
func (item projectItem) Description() string { return item.project.Path }
func (item projectItem) FilterValue() string { return item.project.Name }

// New creates a projectlist model. If preFilter is non-empty, it is applied
// as the initial filter text so the list starts narrowed.
func New(projects []config.Project, preFilter string) Model {
	items := make([]list.Item, len(projects))
	for index, existingProject := range projects {
		items[index] = projectItem{project: existingProject}
	}

	compactDelegate := list.NewDefaultDelegate()
	compactDelegate.ShowDescription = false
	compactDelegate.SetSpacing(0)

	listModel := list.New(items, compactDelegate, 80, 20)
	listModel.Title = "Select a project"

	if preFilter != "" {
		listModel.SetFilteringEnabled(true)
		listModel.SetFilterText(preFilter)
	}

	return Model{list: listModel}
}

// Init implements tea.Model.
func (model Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEnter:
			selectedListItem := model.list.SelectedItem()
			if selectedListItem != nil {
				selected := selectedListItem.(projectItem)
				model.selectedItem = &selected
				return model, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			model.cancelled = true
			return model, tea.Quit
		}
	}

	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		model.list.SetSize(sizeMsg.Width, sizeMsg.Height-4)
	}

	updatedList, listCmd := model.list.Update(msg)
	model.list = updatedList
	return model, listCmd
}

// View implements tea.Model.
func (model Model) View() string {
	return model.list.View()
}

// Result returns the selection outcome after the program exits.
func (model Model) Result() Result {
	if model.cancelled || model.selectedItem == nil {
		return Result{Cancelled: true}
	}
	return Result{
		Project:   model.selectedItem.project,
		Cancelled: false,
	}
}
