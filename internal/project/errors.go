package project

import "errors"

var (
	ErrNotFound      = errors.New("project not found")
	ErrDuplicateName = errors.New("project already exists")
	ErrRelativePath  = errors.New("path must be absolute")
	ErrNoDevScript   = errors.New("project has no dev script")
	ErrNoProjects    = errors.New("no projects configured")
	ErrNameRequired  = errors.New("project name required in non-interactive mode")
)
