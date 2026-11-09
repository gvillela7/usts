package components

import (
	"context"

	"github.com/gvillela7/usts/internal/repository"
)

type projectsCtxKey struct{}

// WithProjects stores the project list used by the sidebar menu.
func WithProjects(ctx context.Context, projects []repository.Project) context.Context {
	return context.WithValue(ctx, projectsCtxKey{}, projects)
}

// ProjectsFromContext returns the project list stored by WithProjects.
func ProjectsFromContext(ctx context.Context) []repository.Project {
	projects, _ := ctx.Value(projectsCtxKey{}).([]repository.Project)
	return projects
}
