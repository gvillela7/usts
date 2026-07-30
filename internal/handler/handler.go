package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/angelofallars/htmx-go"

	"github.com/gvillela7/usts/internal/repository"
	"github.com/gvillela7/usts/web/templates"
	"github.com/gvillela7/usts/web/templates/pages"
	catalogo_templates "github.com/gvillela7/usts/web/templates/pages/catalogo"
	os_templates "github.com/gvillela7/usts/web/templates/pages/os"
	sprints_templates "github.com/gvillela7/usts/web/templates/pages/sprints"
)

type Handler struct {
	Repo repository.Repository
}

func New(repo repository.Repository) *Handler {
	return &Handler{Repo: repo}
}

// IndexViewHandler handles a view for the index page.
func (h *Handler) IndexViewHandler(w http.ResponseWriter, r *http.Request) {
	// Check, if the current URL is '/'.
	if r.URL.Path != "/" {
		// If not, return HTTP 404 error.
		http.NotFound(w, r)
		slog.Error("render page", "method", r.Method, "status", http.StatusNotFound, "path", r.URL.Path)
		return
	}

	// Get projects from repository.
	projects, err := h.Repo.GetProjects()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get projects", "details", err.Error())
		return
	}

	// Define template meta tags.
	metaTags := pages.MetaTags(
		"catálogo, projetos, gestão",         // define meta keywords
		"Bem-vindo ao Catálogo de Projetos.", // define meta description
	)

	// Define template body content.
	bodyContent := pages.BodyContent(
		"Catálogo de Projetos",                                 // define h1 text
		"Selecione um projeto para continuar ou crie um novo.", // define p text
		projects, // pass projects to template
	)

	// Define template layout for index page.
	indexTemplate := templates.Layout(
		"Catálogo de Projetos", // define title text
		"home",                 // define current page
		metaTags,               // define meta tags
		bodyContent,            // define body content
	)

	// Render index page template.
	if err := htmx.NewResponse().RenderTempl(r.Context(), w, indexTemplate); err != nil {
		// Send HTTP 500 error with log.
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}

	// Send log message.
	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

// ShowContentAPIHandler handles an API endpoint to show content.
func (h *Handler) ShowContentAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Check, if the current request has a 'HX-Request' header.
	// For more information, see https://htmx.org/docs/#request-headers
	if !htmx.IsHTMX(r) {
		// If not, return HTTP 400 error.
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		return
	}

	// Write HTML content.
	w.Write([]byte("<p>🎉 Yes, <strong>htmx</strong> is ready to use! (<code>GET /api/hello-world</code>)</p>"))

	// Send htmx response.
	htmx.NewResponse().Write(w)

	// Send log message.
	slog.Info("request API", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

// GetTasksBySprintAPIHandler handles an API endpoint to get tasks by sprint.
func (h *Handler) GetTasksBySprintAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the current request has a 'HX-Request' header
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		return
	}

	sprintIDStr := r.URL.Query().Get("sprint_id")
	if sprintIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Sprint ID is required"))
		return
	}

	var sprintID int
	_, err := fmt.Sscanf(sprintIDStr, "%d", &sprintID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("invalid sprint id", "method", r.Method, "id", sprintIDStr)
		return
	}

	tasks, err := h.Repo.GetTasksBySprint(sprintID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get tasks by sprint", "details", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
	slog.Info("request API", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

// CreateProjectHandler handles project creation via POST request.
func (h *Handler) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the current request has a 'HX-Request' header
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao processar formulário"))
		slog.Error("parse form", "method", r.Method, "status", http.StatusBadRequest, "error", err.Error())
		return
	}

	projectName := r.FormValue("project_name")
	if projectName == "" {
		w.Write([]byte("Nome do projeto é obrigatório"))
		return
	}

	// Create project
	err := h.Repo.CreateProject(projectName)
	if err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("create project", "method", r.Method, "error", err.Error())
		return
	}

	// Success response - return HTML message with redirect header
	w.Header().Set("X-Redirect-After-Success", "/?success=Projeto criado com sucesso")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="text-green-600 font-medium">✓ Projeto criado com sucesso!</div>`))
	slog.Info("project created", "method", r.Method, "status", http.StatusOK, "name", projectName)
}

// CreateSprintHandler handles sprint creation via POST request.
func (h *Handler) CreateSprintHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the current request has a 'HX-Request' header
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao processar formulário"))
		slog.Error("parse form", "method", r.Method, "status", http.StatusBadRequest, "error", err.Error())
		return
	}

	sprintName := r.FormValue("sprint_name")
	projectIDStr := r.FormValue("project_id")

	if sprintName == "" {
		w.Write([]byte("Nome da sprint é obrigatório"))
		return
	}

	if projectIDStr == "" {
		w.Write([]byte("ID do projeto é obrigatório"))
		return
	}

	// Convert projectID to int
	var projectID int
	_, err := fmt.Sscanf(projectIDStr, "%d", &projectID)
	if err != nil {
		w.Write([]byte("ID do projeto inválido"))
		slog.Error("invalid project id", "method", r.Method, "id", projectIDStr)
		return
	}

	// Create sprint
	err = h.Repo.CreateSprint(sprintName, projectID)
	if err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("create sprint", "method", r.Method, "error", err.Error())
		return
	}

	// Success response - return HTML message with redirect header
	w.Header().Set("X-Redirect-After-Success", fmt.Sprintf("/os/new?project_id=%d", projectID))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="text-green-600 font-medium">✓ Sprint criada com sucesso!</div>`))
	slog.Info("sprint created", "method", r.Method, "status", http.StatusOK, "name", sprintName, "project_id", projectID)
}

// CreateTaskHandler handles task creation via POST request.
func (h *Handler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the current request has a 'HX-Request' header
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao processar formulário"))
		slog.Error("parse form", "method", r.Method, "status", http.StatusBadRequest, "error", err.Error())
		return
	}

	taskName := r.FormValue("task_name")
	projectIDStr := r.FormValue("project_id")
	sprintIDStr := r.FormValue("sprint_id")

	if taskName == "" {
		w.Write([]byte("Nome da tarefa é obrigatório"))
		return
	}

	if projectIDStr == "" {
		w.Write([]byte("ID do projeto é obrigatório"))
		return
	}

	// Convert projectID to int
	var projectID int
	_, err := fmt.Sscanf(projectIDStr, "%d", &projectID)
	if err != nil {
		w.Write([]byte("ID do projeto inválido"))
		slog.Error("invalid project id", "method", r.Method, "id", projectIDStr)
		return
	}

	// Convert sprintID to sql.NullInt64
	var sprintID sql.NullInt64
	if sprintIDStr != "" {
		var sid int
		_, err := fmt.Sscanf(sprintIDStr, "%d", &sid)
		if err == nil {
			sprintID = sql.NullInt64{Int64: int64(sid), Valid: true}
		}
	}

	// Create task
	err = h.Repo.CreateTask(taskName, projectID, sprintID)
	if err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("create task", "method", r.Method, "error", err.Error())
		return
	}

	// Success response - return HTML message with redirect header
	w.Header().Set("X-Redirect-After-Success", fmt.Sprintf("/os/new?project_id=%d", projectID))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="text-green-600 font-medium">✓ Tarefa criada com sucesso!</div>`))
	slog.Info("task created", "method", r.Method, "status", http.StatusOK, "name", taskName, "project_id", projectID)
}

// NewOSHandler handles the OS creation form page.
func (h *Handler) NewOSHandler(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("project_id")
	if projectIDStr == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Convert projectID to int
	var projectID int
	_, err := fmt.Sscanf(projectIDStr, "%d", &projectID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("invalid project id", "method", r.Method, "id", projectIDStr)
		return
	}

	// Get project details
	project, err := h.Repo.GetProject(projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		slog.Error("project not found", "method", r.Method, "id", projectID, "error", err.Error())
		return
	}

	// Get catalog list
	catalogs, err := h.Repo.GetCodCatalog()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get catalogs: %v", err), http.StatusInternalServerError)
		return
	}

	// Get sprints for this project
	sprints, err := h.Repo.GetSprints(projectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get sprints: %v", err), http.StatusInternalServerError)
		return
	}

	// Get all tasks for this project (initially empty dropdown, will populate via JS)
	tasks := []repository.Task{}

	// Define template meta tags.
	metaTags := pages.MetaTags(
		"cadastro, os, ordem de serviço",
		"Cadastro de Ordem de Serviço",
	)

	// Define template body content.
	bodyContent := os_templates.Form(project.Name, projectID, sprints, catalogs, tasks)

	// Define template layout for OS form page.
	osTemplate := templates.Layout(
		"Cadastro de OS",
		"projetos",
		metaTags,
		bodyContent,
	)

	// Render OS form template.
	if err := htmx.NewResponse().RenderTempl(r.Context(), w, osTemplate); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}

	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

// OSInput represents the JSON input structure for OS creation
type OSInput struct {
	ProjectID      int     `json:"ProjectID"`
	CatalogID      int     `json:"CatalogID"`
	TaskID         int     `json:"TaskID"`
	SprintID       *int    `json:"SprintID"`
	USTItem        int     `json:"USTItem"`
	USTRecurrence  int     `json:"USTRecurrence"`
	Complexity     int     `json:"Complexity"`
	USTEnd         *int    `json:"USTEnd"`
	Kind           string  `json:"Kind"`
	Description    string  `json:"Description"`
	TarefaCandidta int     `json:"TarefaCandidta"`
}

// CreateOSHandler handles OS creation via POST request (placeholder).
func (h *Handler) CreateOSHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the current request has a 'HX-Request' header
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		return
	}

	// Parse JSON body into intermediate structure
	var inputItems []OSInput
	if err := json.NewDecoder(r.Body).Decode(&inputItems); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao processar dados: " + err.Error()))
		slog.Error("parse JSON", "method", r.Method, "error", err.Error())
		return
	}

	// Validate that at least one item exists
	if len(inputItems) == 0 {
		w.Write([]byte("Adicione pelo menos um item à lista"))
		return
	}

	// Convert input items to repository.OS items
	items := make([]repository.OS, len(inputItems))
	for i, input := range inputItems {
		items[i] = repository.OS{
			ProjectID:      input.ProjectID,
			CatalogID:      input.CatalogID,
			TaskID:         input.TaskID,
			USTItem:        input.USTItem,
			USTRecurrence:  input.USTRecurrence,
			Complexity:     input.Complexity,
			USTEnd:         input.USTEnd,
			Kind:           input.Kind,
			Description:    input.Description,
			TarefaCandidta: input.TarefaCandidta,
		}
		// Convert SprintID from *int to sql.NullInt64
		if input.SprintID != nil {
			items[i].SprintID = sql.NullInt64{Int64: int64(*input.SprintID), Valid: true}
		}
	}

	// Create all OS items
	err := h.Repo.CreateMultipleOS(items)
	if err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("create OS items", "method", r.Method, "error", err.Error())
		return
	}

	// Success response - redirect back to projects page
	w.Header().Set("HX-Redirect", "/?success=OS criada com sucesso")
	w.WriteHeader(http.StatusOK)
	slog.Info("OS created", "method", r.Method, "status", http.StatusOK, "count", len(items))
}

// SprintsViewHandler handles the sprints listing page.
func (h *Handler) SprintsViewHandler(w http.ResponseWriter, r *http.Request) {
	// Get all projects
	projects, err := h.Repo.GetProjects()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get projects", "details", err.Error())
		return
	}

	// Get all sprints
	sprints, err := h.Repo.GetAllSprints()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get sprints", "details", err.Error())
		return
	}

	// Build map of sprint ID to tasks
	sprintTasksMap := make(map[int][]repository.Task)
	for _, sprint := range sprints {
		tasks, err := h.Repo.GetTasksBySprint(sprint.ID)
		if err != nil {
			slog.Warn("failed to get tasks for sprint", "sprint_id", sprint.ID, "error", err.Error())
			tasks = []repository.Task{} // Empty slice on error
		}
		sprintTasksMap[sprint.ID] = tasks
	}

	// Group data by project
	var projectsWithSprints []sprints_templates.ProjectWithSprints
	for _, project := range projects {
		var sprintsWithTasks []sprints_templates.SprintWithTasks
		for _, sprint := range sprints {
			if sprint.ProjectID == project.ID {
				sprintsWithTasks = append(sprintsWithTasks, sprints_templates.SprintWithTasks{
					Sprint: sprint,
					Tasks:  sprintTasksMap[sprint.ID],
				})
			}
		}
		if len(sprintsWithTasks) > 0 {
			projectsWithSprints = append(projectsWithSprints, sprints_templates.ProjectWithSprints{
				Project: project,
				Sprints: sprintsWithTasks,
			})
		}
	}

	// Define template meta tags
	metaTags := sprints_templates.MetaTagsSprints(
		"sprints, gestão, projetos",
		"Visualize todas as sprints organizadas por projeto.",
	)

	// Define template body content with new data structure
	bodyContent := sprints_templates.SprintsBodyContentWithTasks(projectsWithSprints)

	// Define template layout for sprints page
	sprintsTemplate := templates.Layout(
		"Sprints",
		"sprints",
		metaTags,
		bodyContent,
	)

	// Render sprints page template
	if err := htmx.NewResponse().RenderTempl(r.Context(), w, sprintsTemplate); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}

	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

// UpdateSprintHandler handles sprint update via POST request.
func (h *Handler) UpdateSprintHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the current request has a 'HX-Request' header
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao processar formulário"))
		slog.Error("parse form", "method", r.Method, "status", http.StatusBadRequest, "error", err.Error())
		return
	}

	sprintName := r.FormValue("sprint_name")
	sprintIDStr := r.FormValue("sprint_id")

	if sprintName == "" {
		w.Write([]byte("Nome da sprint é obrigatório"))
		return
	}

	if sprintIDStr == "" {
		w.Write([]byte("ID da sprint é obrigatório"))
		return
	}

	// Convert sprintID to int
	var sprintID int
	_, err := fmt.Sscanf(sprintIDStr, "%d", &sprintID)
	if err != nil {
		w.Write([]byte("ID da sprint inválido"))
		slog.Error("invalid sprint id", "method", r.Method, "id", sprintIDStr)
		return
	}

	// Update sprint
	err = h.Repo.UpdateSprint(sprintID, sprintName)
	if err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("update sprint", "method", r.Method, "error", err.Error())
		return
	}

	// Success response - return HTML message with redirect header
	w.Header().Set("X-Redirect-After-Success", "/sprints?success=Sprint atualizada com sucesso")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="text-green-600 font-medium">✓ Sprint atualizada com sucesso!</div>`))
	slog.Info("sprint updated", "method", r.Method, "status", http.StatusOK, "name", sprintName, "sprint_id", sprintID)
}

// UpdateProjectHandler handles project update via POST request.
func (h *Handler) UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the current request has a 'HX-Request' header
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao processar formulário"))
		slog.Error("parse form", "method", r.Method, "status", http.StatusBadRequest, "error", err.Error())
		return
	}

	projectName := r.FormValue("project_name")
	projectIDStr := r.FormValue("project_id")

	if projectName == "" {
		w.Write([]byte("Nome do projeto é obrigatório"))
		return
	}

	if projectIDStr == "" {
		w.Write([]byte("ID do projeto é obrigatório"))
		return
	}

	// Convert projectID to int
	var projectID int
	_, err := fmt.Sscanf(projectIDStr, "%d", &projectID)
	if err != nil {
		w.Write([]byte("ID do projeto inválido"))
		slog.Error("invalid project id", "method", r.Method, "id", projectIDStr)
		return
	}

	// Update project
	err = h.Repo.UpdateProject(projectID, projectName)
	if err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("update project", "method", r.Method, "error", err.Error())
		return
	}

	// Success response - return HTML message with redirect header
	w.Header().Set("X-Redirect-After-Success", "/sprints?success=Projeto atualizado com sucesso")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="text-green-600 font-medium">✓ Projeto atualizado com sucesso!</div>`))
	slog.Info("project updated", "method", r.Method, "status", http.StatusOK, "name", projectName, "project_id", projectID)
}

// TaskDetailViewHandler handles the task detail page showing OS list.
// Note: Route remains /sprints/detail for backward compatibility
func (h *Handler) TaskDetailViewHandler(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("id")
	if taskIDStr == "" {
		http.Redirect(w, r, "/sprints", http.StatusSeeOther)
		return
	}

	// Convert taskID to int
	var taskID int
	_, err := fmt.Sscanf(taskIDStr, "%d", &taskID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("invalid task id", "method", r.Method, "id", taskIDStr)
		return
	}

	// Get task details
	task, err := h.Repo.GetTask(taskID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		slog.Error("task not found", "method", r.Method, "id", taskID, "error", err.Error())
		return
	}

	// Get project details
	project, err := h.Repo.GetProject(task.ProjectID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("project not found", "method", r.Method, "id", task.ProjectID, "error", err.Error())
		return
	}

	// Get OS for this task
	osList, err := h.Repo.GetOSByTask(taskID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get OS by task", "details", err.Error())
		return
	}

	// Define template meta tags
	metaTags := sprints_templates.MetaTagsTaskDetail(
		"task, os, ordem de serviço",
		fmt.Sprintf("Visualize todas as OS da tarefa %s.", task.Name),
	)

	// Define template body content
	bodyContent := sprints_templates.TaskDetailBodyContent(task, project, osList)

	// Define template layout for task detail page
	taskDetailTemplate := templates.Layout(
		fmt.Sprintf("Tarefa: %s", task.Name),
		"sprints",
		metaTags,
		bodyContent,
	)

	// Render task detail page template
	if err := htmx.NewResponse().RenderTempl(r.Context(), w, taskDetailTemplate); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}

	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

// CatalogoViewHandler handles the catalog listing page.
func (h *Handler) CatalogoViewHandler(w http.ResponseWriter, r *http.Request) {
	// Get all catalog items
	catalogs, err := h.Repo.GetCodCatalog()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get catalog", "details", err.Error())
		return
	}

	// Define template meta tags
	metaTags := catalogo_templates.MetaTagsCatalogo(
		"catálogo, serviços, UST",
		"Visualize todos os itens do catálogo de serviços.",
	)

	// Define template body content
	bodyContent := catalogo_templates.CatalogoBodyContent(catalogs)

	// Define template layout for catalog page
	catalogoTemplate := templates.Layout(
		"Catálogo de Serviços",
		"catalogo",
		metaTags,
		bodyContent,
	)

	// Render catalog page template
	if err := htmx.NewResponse().RenderTempl(r.Context(), w, catalogoTemplate); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}

	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}
