package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/angelofallars/htmx-go"
	"github.com/go-chi/chi/v5"

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

func catalogCodLookup(repo repository.Repository, cache map[int]string, catalogID int) string {
	if cod, ok := cache[catalogID]; ok {
		return cod
	}
	cat, err := repo.GetCatalogItem(catalogID)
	if err != nil || cat == nil {
		cache[catalogID] = ""
		return ""
	}
	cod := strings.ToUpper(strings.TrimSpace(cat.Cod))
	cache[catalogID] = cod
	return cod
}

// ENC04 UST mirrors ENC02 Recurrence within the same OS.
func applyENC04DynamicUST(items []repository.OS, codByCatalogID map[int]string) {
	enc02Rec := 0
	enc04Idx := -1
	for i, item := range items {
		switch codByCatalogID[item.CatalogID] {
		case "ENC02":
			rec := item.USTRecurrence
			if rec <= 0 {
				rec = 1
			}
			enc02Rec = rec
		case "ENC04":
			enc04Idx = i
		}
	}
	if enc04Idx < 0 || enc02Rec <= 0 {
		return
	}
	items[enc04Idx].USTItem = enc02Rec
	rec := items[enc04Idx].USTRecurrence
	if rec <= 0 {
		rec = 1
	}
	comp := items[enc04Idx].Complexity
	if comp <= 0 {
		comp = 1
	}
	computed := enc02Rec * rec * comp
	items[enc04Idx].USTEnd = &computed
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
	selected := h.currentProject(w, r)

	// Define template body content.
	bodyContent := pages.BodyContent(
		"Catálogo de Projetos", // define h1 text
		"Selecione um projeto para trabalhar em OS, catálogo e sprints.", // define p text
		projects, // pass projects to template
		selected,
		r.URL.Query().Get("error"),
	)

	// Define template layout for index page.
	indexTemplate := templates.Layout(
		"Catálogo de Projetos", // define title text
		"projetos",             // define current page
		selected,
		metaTags,    // define meta tags
		bodyContent, // define body content
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
// ProjectsListViewHandler renders the page listing all projects.
func (h *Handler) ProjectsListViewHandler(w http.ResponseWriter, r *http.Request) {
	projects, err := h.Repo.GetProjects()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get projects", "details", err.Error())
		return
	}

	current := h.currentProject(w, r)
	page := templates.Layout(
		"Projetos",
		"projetos-lista",
		current,
		pages.MetaTags("projetos, gestão", "Lista de projetos."),
		pages.ProjectsListContent(projects, current),
	)

	if err := htmx.NewResponse().RenderTempl(r.Context(), w, page); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}

	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

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
	w.Header().Set("X-Redirect-After-Success", "/projetos")
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

	sprintName := strings.TrimSpace(r.FormValue("sprint_name"))
	projectIDStr := r.FormValue("project_id")
	startDate, endDate, dateErr := parseSprintDates(r)
	if dateErr != "" {
		w.Write([]byte(dateErr))
		return
	}

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
	err = h.Repo.CreateSprint(sprintName, projectID, startDate, endDate)
	if err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("create sprint", "method", r.Method, "error", err.Error())
		return
	}

	setProjectCookie(w, projectID)
	w.Header().Set("X-Redirect-After-Success", sprintReturnTo(r, projectID))
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
	w.Header().Set("X-Redirect-After-Success", sprintReturnTo(r, projectID))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="text-green-600 font-medium">✓ Tarefa criada com sucesso!</div>`))
	slog.Info("task created", "method", r.Method, "status", http.StatusOK, "name", taskName, "project_id", projectID)
}

// UpdateTaskHandler handles task rename via POST request.
func (h *Handler) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao processar formulário"))
		return
	}

	taskName := strings.TrimSpace(r.FormValue("task_name"))
	taskIDStr := strings.TrimSpace(r.FormValue("task_id"))
	if taskName == "" {
		w.Write([]byte("Nome da tarefa é obrigatório"))
		return
	}
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil || taskID <= 0 {
		w.Write([]byte("ID da tarefa inválido"))
		return
	}

	task, err := h.Repo.GetTask(taskID)
	if err != nil {
		w.Write([]byte("Tarefa não encontrada"))
		return
	}

	if err := h.Repo.UpdateTask(taskID, taskName); err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("update task", "method", r.Method, "error", err.Error())
		return
	}

	setProjectCookie(w, task.ProjectID)
	w.Header().Set("X-Redirect-After-Success", fmt.Sprintf("/sprints?project_id=%d", task.ProjectID))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="text-green-600 font-medium">✓ Tarefa atualizada com sucesso!</div>`))
	slog.Info("task updated", "method", r.Method, "status", http.StatusOK, "task_id", taskID, "name", taskName)
}

// ReorderSprintTasksHandler persists the task order inside a sprint.
func (h *Handler) ReorderSprintTasksHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		SprintID int   `json:"SprintID"`
		TaskIDs  []int `json:"TaskIDs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if payload.SprintID <= 0 || len(payload.TaskIDs) == 0 {
		http.Error(w, "sprint e tarefas são obrigatórios", http.StatusBadRequest)
		return
	}

	sprint, err := h.Repo.GetSprint(payload.SprintID)
	if err != nil || sprint == nil {
		http.Error(w, "Sprint não encontrada", http.StatusNotFound)
		return
	}

	project := h.currentProject(w, r)
	if project == nil || project.ID != sprint.ProjectID {
		http.Error(w, "Sprint não pertence ao projeto atual", http.StatusForbidden)
		return
	}

	if err := h.Repo.ReorderSprintTasks(payload.SprintID, payload.TaskIDs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	slog.Info("tasks reordered", "sprint_id", payload.SprintID, "count", len(payload.TaskIDs))
}

// GetOSByTaskAPIHandler returns OS items for a task (Excel export / clients).
func (h *Handler) GetOSByTaskAPIHandler(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || taskID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID da tarefa inválido"))
		return
	}

	task, err := h.Repo.GetTask(taskID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Tarefa não encontrada"))
		return
	}

	project, err := h.Repo.GetProject(task.ProjectID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Projeto não encontrado"))
		return
	}

	osList, err := h.Repo.GetOSByTask(taskID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Falha ao carregar OS da tarefa"))
		return
	}

	type exportItem struct {
		CatalogCod         string `json:"catalogCod"`
		CatalogDescription string `json:"catalogDescription"`
		Description        string `json:"description"`
		Observation        string `json:"observation"`
		Kind               string `json:"kind"`
		USTItem            int    `json:"ustItem"`
		USTRecurrence      int    `json:"ustRecurrence"`
		Complexity         int    `json:"complexity"`
		USTEnd             int    `json:"ustEnd"`
		TarefaCandidta     int    `json:"tarefaCandidta"`
	}

	items := make([]exportItem, 0, len(osList))
	for _, osItem := range osList {
		ustEnd := 0
		if osItem.OS.USTEnd != nil {
			ustEnd = *osItem.OS.USTEnd
		}
		items = append(items, exportItem{
			CatalogCod:         osItem.CatalogCod,
			CatalogDescription: osItem.CatalogDescription,
			Description:        osItem.OS.Description,
			Observation:        osItem.OS.Observation,
			Kind:               osKindLabel(osItem.OS.Kind),
			USTItem:            osItem.OS.USTItem,
			USTRecurrence:      osItem.OS.USTRecurrence,
			Complexity:         osItem.OS.Complexity,
			USTEnd:             ustEnd,
			TarefaCandidta:     osItem.OS.TarefaCandidta,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"task": map[string]any{
			"id":           task.ID,
			"name":         task.Name,
			"templateKind": repository.NormalizeOSTemplateKind(task.TemplateKind),
		},
		"project": map[string]any{
			"id":   project.ID,
			"name": project.Name,
		},
		"items": items,
	})
}

// GetSprintSummaryAPIHandler returns sprint tasks summary for Excel export.
func (h *Handler) GetSprintSummaryAPIHandler(w http.ResponseWriter, r *http.Request) {
	sprintID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || sprintID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID da sprint inválido"))
		return
	}

	sprint, err := h.Repo.GetSprint(sprintID)
	if err != nil || sprint == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Sprint não encontrada"))
		return
	}

	project, err := h.Repo.GetProject(sprint.ProjectID)
	if err != nil || project == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Projeto da sprint não encontrado"))
		return
	}

	catalogs, err := h.Repo.GetCatalogItemsByProject(project.ID)
	if err != nil {
		catalogs = []repository.Catalog{}
	}

	tasks, err := h.Repo.GetTasksBySprint(sprintID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Falha ao carregar tarefas da sprint"))
		return
	}

	osList, err := h.Repo.GetOSBySprint(sprintID)
	if err != nil {
		osList = []repository.OSWithCatalog{}
	}
	osByTask := make(map[int][]repository.OSWithCatalog)
	for _, item := range osList {
		osByTask[item.OS.TaskID] = append(osByTask[item.OS.TaskID], item)
	}

	type taskRow struct {
		Code         string `json:"code"`
		Responsavel  string `json:"responsavel"`
		Sigla        string `json:"sigla"`
		USTFinal     int    `json:"ustFinal"`
		Observacoes  string `json:"observacoes"`
	}

	rows := make([]taskRow, 0, len(tasks))
	totalUST := 0
	for _, task := range tasks {
		ust := sprints_templates.CalculateTotalUST(osByTask[task.ID], catalogs)
		totalUST += ust
		resp := strings.TrimSpace(task.Responsavel)
		rows = append(rows, taskRow{
			Code:        task.Name,
			Responsavel: resp,
			Sigla:       sprints_templates.ResponsavelInitials(resp),
			USTFinal:    ust,
			Observacoes: "",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"sprint": map[string]any{
			"id":     sprint.ID,
			"title":  sprint.ListTitle(),
			"status": sprint.Status(),
			"name":   sprint.Name,
		},
		"project": map[string]any{
			"id":   project.ID,
			"name": project.Name,
		},
		"tasks":          rows,
		"totalTasks":     len(rows),
		"totalUSTFinal":  totalUST,
	})
}

func osKindLabel(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "nova_tecnologia", "nova tecnologia", "ajuste":
		return "Nova tecnologia"
	default:
		return "Normal"
	}
}

// NewOSHandler handles the OS creation form page.
func (h *Handler) NewOSHandler(w http.ResponseWriter, r *http.Request) {
	project := h.requireProject(w, r, "Selecione um projeto para cadastrar a OS")
	if project == nil {
		return
	}

	catalogs, err := h.Repo.GetCatalogItemsByProject(project.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get catalogs: %v", err), http.StatusInternalServerError)
		return
	}

	sprints, err := h.Repo.GetSprints(project.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get sprints: %v", err), http.StatusInternalServerError)
		return
	}

	tasks := []repository.Task{}
	selectedSprintID := 0
	selectedSprintLabel := ""
	if sprintID, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("sprint_id"))); err == nil && sprintID > 0 {
		sprint, err := h.Repo.GetSprint(sprintID)
		if err == nil && sprint != nil && sprint.ProjectID == project.ID {
			selectedSprintID = sprint.ID
			selectedSprintLabel = sprint.ListTitle()
			tasks, err = h.Repo.GetTasksBySprint(sprint.ID)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to get tasks: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}

	osTmpls, err := h.Repo.GetOSTemplates(project.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get OS templates: %v", err), http.StatusInternalServerError)
		return
	}

	templateKind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("template_kind")))
	if templateKind == repository.OSTemplateKindOS || templateKind == repository.OSTemplateKindEpico {
		filtered := make([]repository.OSTemplate, 0, len(osTmpls))
		for _, tmpl := range osTmpls {
			if tmpl.Kind == templateKind {
				filtered = append(filtered, tmpl)
			}
		}
		osTmpls = filtered
	} else {
		templateKind = ""
	}

	metaTags := pages.MetaTags(
		"cadastro, os, ordem de serviço",
		"Cadastro de Ordem de Serviço",
	)

	bodyContent := os_templates.Form(project.Name, project.ID, sprints, catalogs, tasks, osTmpls, selectedSprintID, selectedSprintLabel, templateKind)

	osTemplate := templates.Layout(
		"Cadastro de OS",
		"os",
		project,
		metaTags,
		bodyContent,
	)

	if err := htmx.NewResponse().RenderTempl(r.Context(), w, osTemplate); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}

	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

// OSInput represents the JSON input structure for OS creation
type OSInput struct {
	ProjectID      int    `json:"ProjectID"`
	CatalogID      int    `json:"CatalogID"`
	TaskID         int    `json:"TaskID"`
	SprintID       *int   `json:"SprintID"`
	USTItem        int    `json:"USTItem"`
	USTRecurrence  int    `json:"USTRecurrence"`
	Complexity     int    `json:"Complexity"`
	USTEnd         *int   `json:"USTEnd"`
	Kind           string `json:"Kind"`
	Description    string `json:"Description"`
	Observation    string `json:"Observation"`
	TarefaCandidta int    `json:"TarefaCandidta"`
}

type CreateOSPayload struct {
	ProjectID    int       `json:"ProjectID"`
	SprintID     int       `json:"SprintID"`
	ActivityCode string    `json:"ActivityCode"`
	Responsavel  string    `json:"Responsavel"`
	TemplateKind string    `json:"TemplateKind"`
	Items        []OSInput `json:"Items"`
}

// CreateOSHandler handles OS creation via POST request (placeholder).
func (h *Handler) CreateOSHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the current request has a 'HX-Request' header
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("request API", "method", r.Method, "status", http.StatusBadRequest, "path", r.URL.Path)
		return
	}

	var payload CreateOSPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao processar dados: " + err.Error()))
		slog.Error("parse JSON", "method", r.Method, "error", err.Error())
		return
	}

	if len(payload.Items) == 0 {
		w.Write([]byte("Adicione pelo menos um item à lista"))
		return
	}
	if payload.ProjectID <= 0 {
		w.Write([]byte("Projeto inválido"))
		return
	}
	if payload.SprintID <= 0 {
		w.Write([]byte("Selecione a sprint"))
		return
	}

	activityCode := strings.TrimSpace(payload.ActivityCode)
	activityCode = strings.ToUpper(activityCode)
	if activityCode == "" || activityCode == "G2MDS-" {
		w.Write([]byte("Informe o código da atividade Jira"))
		return
	}
	if !strings.HasPrefix(activityCode, "G2MDS-") {
		activityCode = "G2MDS-" + strings.TrimPrefix(activityCode, "G2MDS-")
	}
	if len(activityCode) <= len("G2MDS-") {
		w.Write([]byte("Informe o código da atividade Jira"))
		return
	}

	sprint, err := h.Repo.GetSprint(payload.SprintID)
	if err != nil || sprint == nil || sprint.ProjectID != payload.ProjectID {
		w.Write([]byte("Sprint inválida para este projeto"))
		return
	}

	taskID, err := h.Repo.FindOrCreateTask(activityCode, payload.ProjectID, payload.SprintID)
	if err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("find or create task", "method", r.Method, "error", err.Error())
		return
	}

	responsavel := strings.TrimSpace(payload.Responsavel)
	if responsavel == "" {
		w.Write([]byte("Informe o responsável da OS"))
		return
	}
	if err := h.Repo.UpdateTaskResponsavel(taskID, responsavel); err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("update task responsavel", "method", r.Method, "error", err.Error())
		return
	}
	if err := h.Repo.UpdateTaskTemplateKind(taskID, payload.TemplateKind); err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("update task template_kind", "method", r.Method, "error", err.Error())
		return
	}

	sprintID := payload.SprintID
	items := make([]repository.OS, len(payload.Items))
	for i, input := range payload.Items {
		obs := strings.TrimSpace(input.Observation)
		if len([]rune(obs)) > 1000 {
			w.Write([]byte("Observação deve ter no máximo 1000 caracteres"))
			return
		}

		ustItem := input.USTItem
		description := input.Description
		if cat, err := h.Repo.GetCatalogItem(input.CatalogID); err == nil && cat != nil {
			if strings.TrimSpace(description) == "" {
				description = cat.Description
			}
			// If client sent empty/invalid UST, fall back to catalog pricing by kind.
			if ustItem <= 0 {
				ustItem = cat.UstNormal
				if (input.Kind == "nova_tecnologia" || input.Kind == "ajuste") && cat.UstNewTech > 0 {
					ustItem = cat.UstNewTech
				}
			}
		}

		rec := input.USTRecurrence
		if rec <= 0 {
			rec = 1
		}
		complexity := input.Complexity
		if complexity <= 0 {
			complexity = 1
		}
		ustEnd := input.USTEnd
		computed := ustItem * rec * complexity
		if ustEnd == nil || *ustEnd <= 0 {
			ustEnd = &computed
		}

		items[i] = repository.OS{
			ProjectID:      payload.ProjectID,
			CatalogID:      input.CatalogID,
			TaskID:         taskID,
			SprintID:       sql.NullInt64{Int64: int64(sprintID), Valid: true},
			USTItem:        ustItem,
			USTRecurrence:  rec,
			Complexity:     complexity,
			USTEnd:         ustEnd,
			Kind:           input.Kind,
			Description:    description,
			Observation:    obs,
			SortOrder:      i + 1,
			TarefaCandidta: input.TarefaCandidta,
		}
	}

	catalogCodCache := make(map[int]string)
	codByCatalogID := make(map[int]string, len(items))
	for _, item := range items {
		if _, ok := codByCatalogID[item.CatalogID]; !ok {
			codByCatalogID[item.CatalogID] = catalogCodLookup(h.Repo, catalogCodCache, item.CatalogID)
		}
	}
	applyENC04DynamicUST(items, codByCatalogID)

	err = h.Repo.CreateMultipleOS(items)
	if err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("create OS items", "method", r.Method, "error", err.Error())
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/sprints?project_id=%d", payload.ProjectID))
	w.WriteHeader(http.StatusOK)
	slog.Info("OS created", "method", r.Method, "status", http.StatusOK, "count", len(items), "project_id", payload.ProjectID, "activity", activityCode)
}

// SprintsViewHandler handles the sprints listing page.
func (h *Handler) SprintsViewHandler(w http.ResponseWriter, r *http.Request) {
	project := h.requireProject(w, r, "Selecione um projeto para ver as sprints")
	if project == nil {
		return
	}

	sprints, err := h.Repo.GetSprints(project.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get sprints", "details", err.Error())
		return
	}

	catalogs, err := h.Repo.GetCatalogItemsByProject(project.ID)
	if err != nil {
		slog.Warn("failed to get catalogs for sprint UST totals", "project_id", project.ID, "error", err.Error())
		catalogs = []repository.Catalog{}
	}

	var sprintsWithTasks []sprints_templates.SprintWithTasks
	for _, sprint := range sprints {
		tasks, err := h.Repo.GetTasksBySprint(sprint.ID)
		if err != nil {
			slog.Warn("failed to get tasks for sprint", "sprint_id", sprint.ID, "error", err.Error())
			tasks = []repository.Task{}
		}

		osList, err := h.Repo.GetOSBySprint(sprint.ID)
		if err != nil {
			slog.Warn("failed to get OS for sprint UST totals", "sprint_id", sprint.ID, "error", err.Error())
			osList = []repository.OSWithCatalog{}
		}
		osByTask := make(map[int][]repository.OSWithCatalog)
		for _, item := range osList {
			osByTask[item.OS.TaskID] = append(osByTask[item.OS.TaskID], item)
		}

		tasksWithMeta := make([]sprints_templates.TaskWithMeta, 0, len(tasks))
		sprintUSTTotal := 0
		for _, task := range tasks {
			taskUST := sprints_templates.CalculateTotalUST(osByTask[task.ID], catalogs)
			sprintUSTTotal += taskUST
			tasksWithMeta = append(tasksWithMeta, sprints_templates.TaskWithMeta{
				Task:          task,
				USTFinalTotal: taskUST,
			})
		}

		sprintsWithTasks = append(sprintsWithTasks, sprints_templates.SprintWithTasks{
			Sprint:         sprint,
			Tasks:          tasksWithMeta,
			USTFinalTotal:  sprintUSTTotal,
		})
	}

	metaTags := sprints_templates.MetaTagsSprints(
		"sprints, gestão, projetos",
		"Visualize as sprints do projeto.",
	)

	bodyContent := sprints_templates.SprintsBodyContentWithTasks(*project, sprintsWithTasks)

	sprintsTemplate := templates.Layout(
		"Sprints",
		"sprints",
		project,
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

	sprintName := strings.TrimSpace(r.FormValue("sprint_name"))
	sprintIDStr := r.FormValue("sprint_id")
	startDate, endDate, dateErr := parseSprintDates(r)
	if dateErr != "" {
		w.Write([]byte(dateErr))
		return
	}

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
	err = h.Repo.UpdateSprint(sprintID, sprintName, startDate, endDate)
	if err != nil {
		w.Write([]byte(err.Error()))
		slog.Error("update sprint", "method", r.Method, "error", err.Error())
		return
	}

	redirect := "/sprints"
	if sprint, sprintErr := h.Repo.GetSprint(sprintID); sprintErr == nil {
		setProjectCookie(w, sprint.ProjectID)
		redirect = fmt.Sprintf("/sprints?project_id=%d", sprint.ProjectID)
	}
	w.Header().Set("X-Redirect-After-Success", redirect)
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

	setProjectCookie(w, projectID)
	w.Header().Set("X-Redirect-After-Success", fmt.Sprintf("/sprints?project_id=%d", projectID))
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
	setProjectCookie(w, project.ID)

	// Get OS for this task
	osList, err := h.Repo.GetOSByTask(taskID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get OS by task", "details", err.Error())
		return
	}

	catalogs, err := h.Repo.GetCatalogItemsByProject(project.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get catalogs", "details", err.Error())
		return
	}

	editMode := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("edit")), "1") ||
		strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("edit")), "true")

	// Define template meta tags
	metaTags := sprints_templates.MetaTagsTaskDetail(
		"task, os, ordem de serviço",
		fmt.Sprintf("Visualize todas as OS da tarefa %s.", task.Name),
	)

	// Define template body content
	bodyContent := sprints_templates.TaskDetailBodyContent(task, project, osList, catalogs, editMode)

	// Define template layout for task detail page
	taskDetailTemplate := templates.Layout(
		fmt.Sprintf("Tarefa: %s", task.Name),
		"sprints",
		project,
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

// SaveTaskOSHandler saves edited OS items for a task (update/delete/reorder).
func (h *Handler) SaveTaskOSHandler(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || taskID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ID da tarefa inválido"))
		return
	}

	task, err := h.Repo.GetTask(taskID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Tarefa não encontrada"))
		return
	}

	var payload struct {
		Responsavel string `json:"Responsavel"`
		Items       []struct {
			ID             int    `json:"ID"`
			CatalogID      int    `json:"CatalogID"`
			USTItem        int    `json:"USTItem"`
			USTRecurrence  int    `json:"USTRecurrence"`
			Complexity     int    `json:"Complexity"`
			USTEnd         *int   `json:"USTEnd"`
			Kind           string `json:"Kind"`
			Description    string `json:"Description"`
			Observation    string `json:"Observation"`
			TarefaCandidta int    `json:"TarefaCandidta"`
		} `json:"Items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Erro ao processar dados: " + err.Error()))
		return
	}

	responsavel := strings.TrimSpace(payload.Responsavel)
	if responsavel == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Informe o responsável da OS"))
		return
	}
	if err := h.Repo.UpdateTaskResponsavel(task.ID, responsavel); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	items := make([]repository.OS, 0, len(payload.Items))
	for _, in := range payload.Items {
		if in.CatalogID <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Cada item precisa de um código do catálogo"))
			return
		}
		obs := strings.TrimSpace(in.Observation)
		if len([]rune(obs)) > 1000 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Observação deve ter no máximo 1000 caracteres"))
			return
		}
		rec := in.USTRecurrence
		if rec <= 0 {
			rec = 1
		}
		complexity := in.Complexity
		if complexity <= 0 {
			complexity = 1
		}
		ustItem := in.USTItem
		if cat, err := h.Repo.GetCatalogItem(in.CatalogID); err == nil && cat != nil {
			cod := strings.ToUpper(strings.TrimSpace(cat.Cod))
			if cod != "ENC04" {
				k := strings.ToLower(strings.TrimSpace(in.Kind))
				if k == "nova_tecnologia" || k == "ajuste" {
					if cat.UstNewTech > 0 {
						ustItem = cat.UstNewTech
					} else if cat.UstNormal > 0 {
						ustItem = cat.UstNormal
					}
				} else if cat.UstNormal > 0 {
					ustItem = cat.UstNormal
				}
			}
		}
		ustEnd := in.USTEnd
		computed := ustItem * rec * complexity
		if ustEnd == nil || *ustEnd <= 0 {
			ustEnd = &computed
		}
		description := strings.TrimSpace(in.Description)
		if cat, err := h.Repo.GetCatalogItem(in.CatalogID); err == nil && cat != nil {
			if description == "" {
				description = cat.Description
			}
		}

		items = append(items, repository.OS{
			ID:             in.ID,
			ProjectID:      task.ProjectID,
			CatalogID:      in.CatalogID,
			TaskID:         task.ID,
			SprintID:       task.SprintID,
			USTItem:        ustItem,
			USTRecurrence:  rec,
			Complexity:     complexity,
			USTEnd:         ustEnd,
			Kind:           in.Kind,
			Description:    description,
			Observation:    obs,
			TarefaCandidta: in.TarefaCandidta,
		})
	}

	catalogCodCache := make(map[int]string)
	codByCatalogID := make(map[int]string, len(items))
	for _, item := range items {
		if _, ok := codByCatalogID[item.CatalogID]; !ok {
			codByCatalogID[item.CatalogID] = catalogCodLookup(h.Repo, catalogCodCache, item.CatalogID)
		}
	}
	applyENC04DynamicUST(items, codByCatalogID)

	if err := h.Repo.SaveTaskOSItems(taskID, items); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		slog.Error("save task OS", "error", err.Error(), "task_id", taskID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"redirect": fmt.Sprintf("/sprints/detail?id=%d", taskID),
	})
	slog.Info("task OS saved", "task_id", taskID, "count", len(items))
}

// CatalogoViewHandler handles the catalog listing page.
func (h *Handler) CatalogoViewHandler(w http.ResponseWriter, r *http.Request) {
	project := h.requireProject(w, r, "Selecione um projeto para ver o catálogo")
	if project == nil {
		return
	}

	catalogs, err := h.Repo.GetServiceCatalogsByProject(project.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get catalogs", "details", err.Error())
		return
	}

	metaTags := catalogo_templates.MetaTagsCatalogo(
		"catálogo, serviços, UST",
		"Cadastre o catálogo de serviços do projeto.",
	)

	bodyContent := catalogo_templates.CatalogoBodyContent(catalogs, *project)
	catalogoTemplate := templates.Layout(
		"Catálogo de Serviços",
		"catalogo",
		project,
		metaTags,
		bodyContent,
	)

	if err := htmx.NewResponse().RenderTempl(r.Context(), w, catalogoTemplate); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}

	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

func (h *Handler) CatalogoNewViewHandler(w http.ResponseWriter, r *http.Request) {
	project := h.requireProject(w, r, "Selecione um projeto para cadastrar o catálogo")
	if project == nil {
		return
	}

	hasTemplates, err := h.Repo.HasTemplateCatalogItems()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to check template catalog items", "details", err.Error())
		return
	}

	now := time.Now()
	defaultStart := now.Format("2006-01-02")
	defaultEnd := now.AddDate(1, 0, 0).Format("2006-01-02")

	metaTags := catalogo_templates.MetaTagsCatalogo(
		"catálogo, cadastro, serviços",
		"Cadastre um catálogo de serviços para o projeto.",
	)
	bodyContent := catalogo_templates.CatalogoNewBodyContent(
		*project,
		hasTemplates,
		defaultStart,
		defaultEnd,
		r.URL.Query().Get("error"),
	)
	page := templates.Layout("Novo catálogo", "catalogo", project, metaTags, bodyContent)
	if err := htmx.NewResponse().RenderTempl(r.Context(), w, page); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}
	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

func (h *Handler) CatalogoDetailViewHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(r.URL.Query().Get("id"))
	if !ok {
		http.Redirect(w, r, "/catalogo", http.StatusSeeOther)
		return
	}

	catalog, err := h.Repo.GetServiceCatalog(id)
	if err != nil {
		slog.Error("failed to load catalog", "method", r.Method, "id", id, "error", err.Error())
		http.Error(w, "Não foi possível abrir o catálogo.", http.StatusInternalServerError)
		return
	}

	project, err := h.Repo.GetProject(catalog.ProjectID)
	if err != nil {
		http.Redirect(w, r, "/catalogo", http.StatusSeeOther)
		return
	}
	setProjectCookie(w, project.ID)

	items, err := h.Repo.GetCatalogItems(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to get catalog items", "details", err.Error())
		return
	}

	metaTags := catalogo_templates.MetaTagsCatalogo(
		"catálogo, serviços, UST",
		fmt.Sprintf("Itens do catálogo %s.", catalog.Name),
	)
	bodyContent := catalogo_templates.CatalogoDetailBodyContent(*catalog, items, r.URL.Query().Get("error"))
	page := templates.Layout(catalog.Name, "catalogo", project, metaTags, bodyContent)
	if err := htmx.NewResponse().RenderTempl(r.Context(), w, page); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}
	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

func (h *Handler) CreateCatalogHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape("Erro ao processar formulário"), http.StatusSeeOther)
		return
	}

	projectID, ok := parseIDParam(r.FormValue("project_id"))
	if !ok {
		http.Redirect(w, r, "/?error="+url.QueryEscape("Selecione um projeto para cadastrar o catálogo"), http.StatusSeeOther)
		return
	}
	redirectErr := func(message string) {
		http.Redirect(w, r, fmt.Sprintf("/catalogo/new?project_id=%d&error=%s", projectID, url.QueryEscape(message)), http.StatusSeeOther)
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		redirectErr("Nome do catálogo é obrigatório")
		return
	}

	startDate, err := time.Parse("2006-01-02", r.FormValue("start_date"))
	if err != nil {
		redirectErr("Data de início inválida")
		return
	}
	endDate, err := time.Parse("2006-01-02", r.FormValue("end_date"))
	if err != nil {
		redirectErr("Data de fim inválida")
		return
	}
	if endDate.Before(startDate) {
		redirectErr("A data de fim deve ser posterior ao início")
		return
	}

	if _, err := h.Repo.GetProject(projectID); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape("Projeto não encontrado"), http.StatusSeeOther)
		return
	}
	setProjectCookie(w, projectID)

	catalogID, err := h.Repo.CreateServiceCatalog(projectID, name, startDate, endDate)
	if err != nil {
		redirectErr("Não foi possível criar o catálogo")
		slog.Error("create catalog", "error", err.Error())
		return
	}

	if r.FormValue("copy_template") == "1" {
		if err := h.Repo.CopyTemplateItems(catalogID); err != nil {
			slog.Error("copy template catalog items", "error", err.Error(), "catalog_id", catalogID)
		}
	}

	slog.Info("catalog created", "id", catalogID, "name", name, "project_id", projectID)
	http.Redirect(w, r, fmt.Sprintf("/catalogo/detail?id=%d", catalogID), http.StatusSeeOther)
}

func (h *Handler) CreateCatalogItemHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/catalogo", http.StatusSeeOther)
		return
	}

	catalogID, ok := parseIDParam(r.FormValue("catalog_id"))
	if !ok {
		http.Redirect(w, r, "/catalogo", http.StatusSeeOther)
		return
	}

	redirectErr := func(message string) {
		http.Redirect(w, r, fmt.Sprintf("/catalogo/detail?id=%d&error=%s", catalogID, url.QueryEscape(message)), http.StatusSeeOther)
	}

	cod, description, ustNormal, ustNewTech, errMsg := parseCatalogItemForm(r)
	if errMsg != "" {
		redirectErr(errMsg)
		return
	}

	if _, err := h.Repo.GetServiceCatalog(catalogID); err != nil {
		http.Redirect(w, r, "/catalogo", http.StatusSeeOther)
		return
	}

	if err := h.Repo.CreateCatalogItem(catalogID, strings.ToUpper(cod), description, ustNormal, ustNewTech, "Unidade"); err != nil {
		redirectErr("Não foi possível incluir o item")
		slog.Error("create catalog item", "error", err.Error())
		return
	}

	slog.Info("catalog item created", "catalog_id", catalogID, "cod", cod)
	http.Redirect(w, r, fmt.Sprintf("/catalogo/detail?id=%d", catalogID), http.StatusSeeOther)
}

func (h *Handler) UpdateCatalogItemHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/catalogo", http.StatusSeeOther)
		return
	}

	catalogID, ok := parseIDParam(r.FormValue("catalog_id"))
	if !ok {
		http.Redirect(w, r, "/catalogo", http.StatusSeeOther)
		return
	}

	redirectErr := func(message string) {
		http.Redirect(w, r, fmt.Sprintf("/catalogo/detail?id=%d&error=%s", catalogID, url.QueryEscape(message)), http.StatusSeeOther)
	}

	itemID, ok := parseIDParam(r.FormValue("item_id"))
	if !ok {
		redirectErr("Selecione o item que deseja editar")
		return
	}

	cod := strings.TrimSpace(r.FormValue("cod"))
	description := strings.TrimSpace(r.FormValue("description"))
	if cod == "" || description == "" {
		redirectErr("Código e descrição são obrigatórios")
		return
	}

	ustNormal, err := strconv.Atoi(strings.TrimSpace(r.FormValue("ust_normal")))
	if err != nil || ustNormal < 0 {
		redirectErr("UST Normal inválido")
		return
	}

	var ustNewTech *int
	if raw := strings.TrimSpace(r.FormValue("ust_new_tech")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			redirectErr("UST New Tech inválido")
			return
		}
		ustNewTech = &value
	}

	item, err := h.Repo.GetCatalogItem(itemID)
	if err != nil || item.CatalogID != catalogID {
		redirectErr("Item não encontrado neste catálogo")
		return
	}

	if err := h.Repo.UpdateCatalogItem(itemID, catalogID, strings.ToUpper(cod), description, ustNormal, ustNewTech); err != nil {
		redirectErr("Não foi possível salvar o item")
		slog.Error("update catalog item", "error", err.Error(), "item_id", itemID)
		return
	}

	slog.Info("catalog item updated", "catalog_id", catalogID, "item_id", itemID, "cod", cod)
	http.Redirect(w, r, fmt.Sprintf("/catalogo/detail?id=%d", catalogID), http.StatusSeeOther)
}

func parseCatalogItemForm(r *http.Request) (cod, description string, ustNormal int, ustNewTech *int, errMsg string) {
	cod = strings.TrimSpace(r.FormValue("cod"))
	description = strings.TrimSpace(r.FormValue("description"))
	if cod == "" || description == "" {
		return "", "", 0, nil, "Código e descrição são obrigatórios"
	}

	ustNormal, err := strconv.Atoi(strings.TrimSpace(r.FormValue("ust_normal")))
	if err != nil || ustNormal < 0 {
		return "", "", 0, nil, "UST Normal inválido"
	}

	if raw := strings.TrimSpace(r.FormValue("ust_new_tech")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return "", "", 0, nil, "UST New Tech inválido"
		}
		ustNewTech = &value
	}
	return cod, description, ustNormal, ustNewTech, ""
}

func parseIDParam(value string) (int, bool) {
	if strings.TrimSpace(value) == "" {
		return 0, false
	}
	var id int
	if _, err := fmt.Sscanf(value, "%d", &id); err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

const projectCookieName = "usts_project_id"

func setProjectCookie(w http.ResponseWriter, projectID int) {
	http.SetCookie(w, &http.Cookie{
		Name:     projectCookieName,
		Value:    strconv.Itoa(projectID),
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) currentProject(w http.ResponseWriter, r *http.Request) *repository.Project {
	idStr := r.URL.Query().Get("project_id")
	if idStr == "" {
		if cookie, err := r.Cookie(projectCookieName); err == nil {
			idStr = cookie.Value
		}
	}
	id, ok := parseIDParam(idStr)
	if !ok {
		return nil
	}
	project, err := h.Repo.GetProject(id)
	if err != nil {
		return nil
	}
	setProjectCookie(w, project.ID)
	return project
}

func (h *Handler) requireProject(w http.ResponseWriter, r *http.Request, message string) *repository.Project {
	project := h.currentProject(w, r)
	if project == nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape(message), http.StatusSeeOther)
	}
	return project
}

func parseSprintDates(r *http.Request) (time.Time, time.Time, string) {
	startRaw := strings.TrimSpace(r.FormValue("start_date"))
	endRaw := strings.TrimSpace(r.FormValue("end_date"))
	if startRaw == "" || endRaw == "" {
		return time.Time{}, time.Time{}, "Informe a data de início e a data de fim"
	}
	startDate, err := time.Parse("2006-01-02", startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, "Data de início inválida"
	}
	endDate, err := time.Parse("2006-01-02", endRaw)
	if err != nil {
		return time.Time{}, time.Time{}, "Data de fim inválida"
	}
	if !endDate.After(startDate) {
		return time.Time{}, time.Time{}, "A data de fim deve ser posterior à data de início. Não pode ser o mesmo dia."
	}
	return startDate, endDate, ""
}

func sprintReturnTo(r *http.Request, projectID int) string {
	fallback := fmt.Sprintf("/os/new?project_id=%d", projectID)
	returnTo := strings.TrimSpace(r.FormValue("return_to"))
	if returnTo == "" {
		return fallback
	}
	if strings.HasPrefix(returnTo, "/sprints") || strings.HasPrefix(returnTo, "/os/new") {
		if !strings.Contains(returnTo, "project_id=") {
			sep := "?"
			if strings.Contains(returnTo, "?") {
				sep = "&"
			}
			returnTo = fmt.Sprintf("%s%sproject_id=%d", returnTo, sep, projectID)
		}
		return returnTo
	}
	return fallback
}

func (h *Handler) OSTemplatesViewHandler(w http.ResponseWriter, r *http.Request) {
	project := h.requireProject(w, r, "Selecione um projeto para gerenciar templates de OS")
	if project == nil {
		return
	}

	list, err := h.Repo.GetOSTemplates(project.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get OS templates: %v", err), http.StatusInternalServerError)
		return
	}

	metaTags := pages.MetaTags("templates, os", "Templates de Ordem de Serviço")
	bodyContent := os_templates.TemplatesList(*project, list)
	page := templates.Layout("Templates de OS", "os-templates", project, metaTags, bodyContent)

	if err := htmx.NewResponse().RenderTempl(r.Context(), w, page); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}
	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

func (h *Handler) OSTemplateNewViewHandler(w http.ResponseWriter, r *http.Request) {
	project := h.requireProject(w, r, "Selecione um projeto para criar template de OS")
	if project == nil {
		return
	}

	catalogs, err := h.Repo.GetCatalogItemsByProject(project.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get catalogs: %v", err), http.StatusInternalServerError)
		return
	}

	metaTags := pages.MetaTags("templates, os, novo", "Novo template de OS")
	bodyContent := os_templates.TemplateForm(project.Name, project.ID, catalogs, nil, nil)
	page := templates.Layout("Novo template de OS", "os-templates", project, metaTags, bodyContent)

	if err := htmx.NewResponse().RenderTempl(r.Context(), w, page); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}
	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

func (h *Handler) OSTemplateEditViewHandler(w http.ResponseWriter, r *http.Request) {
	project := h.requireProject(w, r, "Selecione um projeto para editar template de OS")
	if project == nil {
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "template id inválido", http.StatusBadRequest)
		return
	}

	tmpl, err := h.Repo.GetOSTemplate(id)
	if err != nil || tmpl.ProjectID != project.ID {
		http.Error(w, "template não encontrado", http.StatusNotFound)
		return
	}

	items, err := h.Repo.GetOSTemplateItems(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get template items: %v", err), http.StatusInternalServerError)
		return
	}

	catalogs, err := h.Repo.GetCatalogItemsByProject(project.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get catalogs: %v", err), http.StatusInternalServerError)
		return
	}

	metaTags := pages.MetaTags("templates, os, editar", "Editar template de OS")
	bodyContent := os_templates.TemplateForm(project.Name, project.ID, catalogs, tmpl, items)
	page := templates.Layout("Editar template de OS", "os-templates", project, metaTags, bodyContent)

	if err := htmx.NewResponse().RenderTempl(r.Context(), w, page); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("render template", "method", r.Method, "status", http.StatusInternalServerError, "path", r.URL.Path)
		return
	}
	slog.Info("render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

type OSTemplatePayload struct {
	ID        int                         `json:"ID"`
	ProjectID int                         `json:"ProjectID"`
	Name      string                      `json:"Name"`
	Kind      string                      `json:"Kind"`
	Items     []repository.OSTemplateItem `json:"Items"`
}

func (h *Handler) CreateOSTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload OSTemplatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Erro ao processar dados: "+err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(payload.Name)
	if name == "" || payload.ProjectID <= 0 || len(payload.Items) == 0 {
		http.Error(w, "Nome, projeto e pelo menos um item são obrigatórios", http.StatusBadRequest)
		return
	}

	if _, err := h.Repo.GetProject(payload.ProjectID); err != nil {
		http.Error(w, "Projeto não encontrado", http.StatusBadRequest)
		return
	}

	id, err := h.Repo.CreateOSTemplate(payload.ProjectID, name, payload.Kind, payload.Items)
	if err != nil {
		http.Error(w, "Erro ao criar template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]int{"id": id})
	slog.Info("OS template created", "id", id, "project_id", payload.ProjectID)
}

func (h *Handler) UpdateOSTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if !htmx.IsHTMX(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload OSTemplatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Erro ao processar dados: "+err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(payload.Name)
	if payload.ID <= 0 || name == "" || payload.ProjectID <= 0 || len(payload.Items) == 0 {
		http.Error(w, "Dados inválidos para atualizar o template", http.StatusBadRequest)
		return
	}

	existing, err := h.Repo.GetOSTemplate(payload.ID)
	if err != nil || existing.ProjectID != payload.ProjectID {
		http.Error(w, "Template não encontrado", http.StatusNotFound)
		return
	}

	if err := h.Repo.UpdateOSTemplate(payload.ID, payload.ProjectID, name, payload.Kind, payload.Items); err != nil {
		http.Error(w, "Erro ao atualizar template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
	slog.Info("OS template updated", "id", payload.ID, "project_id", payload.ProjectID)
}

func (h *Handler) DeleteOSTemplateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	projectID, err := strconv.Atoi(r.FormValue("project_id"))
	if err != nil || projectID <= 0 {
		http.Error(w, "project_id inválido", http.StatusBadRequest)
		return
	}

	if err := h.Repo.DeleteOSTemplate(id, projectID); err != nil {
		http.Error(w, "Erro ao excluir template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/os/templates?project_id=%d", projectID), http.StatusSeeOther)
}

func (h *Handler) GetOSTemplateItemsAPIHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	tmpl, err := h.Repo.GetOSTemplate(id)
	if err != nil {
		http.Error(w, "template não encontrado", http.StatusNotFound)
		return
	}

	project := h.currentProject(w, r)
	if project == nil || project.ID != tmpl.ProjectID {
		http.Error(w, "template não pertence ao projeto atual", http.StatusForbidden)
		return
	}

	items, err := h.Repo.GetOSTemplateItems(id)
	if err != nil {
		http.Error(w, "failed to get items", http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []repository.OSTemplateItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"kind":  repository.NormalizeOSTemplateKind(tmpl.Kind),
		"items": items,
	}); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}
}
