package project

import (
	"path/filepath"
	"strings"

	"github.com/ArielSurco/cli/internal/config"
)

// Service handles project business logic.
// It never calls config.Save() — that is the responsibility of the cmd layer.
type Service struct {
	cfg *config.Config
}

// NewService creates a new Service backed by the given config.
func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

// Add registers a new project. Returns ErrDuplicateName if name already exists,
// ErrRelativePath if path is not absolute.
func (svc *Service) Add(name, path, devScript string) error {
	if !filepath.IsAbs(path) {
		return ErrRelativePath
	}
	for _, existingProject := range svc.cfg.Projects {
		if existingProject.Name == name {
			return ErrDuplicateName
		}
	}
	svc.cfg.Projects = append(svc.cfg.Projects, config.Project{
		Name:      name,
		Path:      path,
		DevScript: devScript,
	})
	return nil
}

// Remove deletes the project with the given name. Returns ErrNotFound if not present.
func (svc *Service) Remove(name string) error {
	for index, existingProject := range svc.cfg.Projects {
		if existingProject.Name == name {
			svc.cfg.Projects = append(svc.cfg.Projects[:index], svc.cfg.Projects[index+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// FindByName returns the project with the given name or ErrNotFound.
func (svc *Service) FindByName(name string) (*config.Project, error) {
	for index := range svc.cfg.Projects {
		if svc.cfg.Projects[index].Name == name {
			return &svc.cfg.Projects[index], nil
		}
	}
	return nil, ErrNotFound
}

// List returns all configured projects.
func (svc *Service) List() []config.Project {
	return svc.cfg.Projects
}

// DevCommand returns the shell command to cd into the project and run its dev script.
// Single-quote escapes the path for POSIX shell safety.
// Returns ErrNotFound if the project does not exist, ErrNoDevScript if no script is set.
func (svc *Service) DevCommand(name string) (string, error) {
	foundProject, err := svc.FindByName(name)
	if err != nil {
		return "", err
	}
	if foundProject.DevScript == "" {
		return "", ErrNoDevScript
	}
	escapedPath := strings.ReplaceAll(foundProject.Path, "'", `'\''`)
	return "cd '" + escapedPath + "' && " + foundProject.DevScript, nil
}
