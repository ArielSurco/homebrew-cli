package module

// CommandDef defines a shell alias and its backing Cobra subcommand.
type CommandDef struct {
	Alias          string // shell function name, e.g. "gp"
	CobraCmd       string // cobra subcommand path, e.g. "project go"
	NeedsEval      bool   // true = wrap in eval "$(...)"; false = call directly
	CdOutput       bool   // true = capture output and cd into it (e.g. gp)
	HasCompletions bool   // true = emit completion wiring (bash + zsh)
	Description    string
}

// Module groups related commands under a named module.
type Module struct {
	Name     string
	Commands []CommandDef
}

// Registry is the compiled-in list of all available modules.
// To add a new module: append a Module entry here and add its Cobra commands.
var Registry = []Module{
	{
		Name: "go-project",
		Commands: []CommandDef{
			{Alias: "gp", CobraCmd: "project go", NeedsEval: false, CdOutput: true, HasCompletions: true, Description: "Navigate to a project directory"},
			{Alias: "gpd", CobraCmd: "project dev", NeedsEval: true, HasCompletions: true, Description: "Run the dev script for a project"},
			{Alias: "gpa", CobraCmd: "project add", NeedsEval: false, HasCompletions: false, Description: "Register a new project"},
			{Alias: "gpr", CobraCmd: "project remove", NeedsEval: false, HasCompletions: false, Description: "Unregister a project"},
		},
	},
}

// FindModule returns the module with the given name, if it exists.
func FindModule(name string) (*Module, bool) {
	for index := range Registry {
		if Registry[index].Name == name {
			return &Registry[index], true
		}
	}
	return nil, false
}

// ActiveModules returns modules from Registry whose names are in activeNames.
// Order follows Registry order, not activeNames order.
func ActiveModules(activeNames []string) []Module {
	nameSet := make(map[string]bool, len(activeNames))
	for _, activeName := range activeNames {
		nameSet[activeName] = true
	}

	result := make([]Module, 0)
	for _, registryModule := range Registry {
		if nameSet[registryModule.Name] {
			result = append(result, registryModule)
		}
	}
	return result
}
