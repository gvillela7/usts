package sprints

import "github.com/gvillela7/usts/internal/repository"

// TaskWithMeta extends a task with computed UST Final total.
type TaskWithMeta struct {
	Task          repository.Task
	USTFinalTotal int
}

// SprintWithTasks groups a sprint with its associated tasks
type SprintWithTasks struct {
	Sprint         repository.Sprint
	Tasks          []TaskWithMeta
	USTFinalTotal  int
}

// ProjectWithSprints groups a project with its sprints and tasks
type ProjectWithSprints struct {
	Project repository.Project
	Sprints []SprintWithTasks
}
