package sprints

import "github.com/gvillela7/usts/internal/repository"

// SprintWithTasks groups a sprint with its associated tasks
type SprintWithTasks struct {
	Sprint repository.Sprint
	Tasks  []repository.Task
}

// ProjectWithSprints groups a project with its sprints and tasks
type ProjectWithSprints struct {
	Project repository.Project
	Sprints []SprintWithTasks
}
