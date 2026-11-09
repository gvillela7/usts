package routes

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	gowebly "github.com/gowebly/helpers"
	handler "github.com/gvillela7/usts/internal/handler"
	"github.com/gvillela7/usts/internal/repository"
	"github.com/gvillela7/usts/web/templates/components"
)

func Setup(db *sql.DB, static http.FileSystem) *chi.Mux {
	// Create a new chi router.
	router := chi.NewRouter()

	// Use chi middlewares.

	router.Use(middleware.Logger)

	// Initialize repository and handlers.
	repo := repository.NewSQLiteRepository(db)
	handler := handler.New(repo)

	// Load the project list for the sidebar menu on page requests.
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && !strings.HasPrefix(r.URL.Path, "/static/") && !strings.HasPrefix(r.URL.Path, "/api/") {
				if projects, err := repo.GetProjects(); err == nil {
					r = r.WithContext(components.WithProjects(r.Context(), projects))
				}
			}
			next.ServeHTTP(w, r)
		})
	})

	// Handle static files from the embed FS (with a custom handler).
	router.Handle("/static/*", gowebly.StaticFileServerHandler(static))

	// Handle index page view.
	router.Get("/", handler.IndexViewHandler)
	router.Get("/projetos", handler.ProjectsListViewHandler)

	// Handle catalog page view.
	router.Get("/catalogo", handler.CatalogoViewHandler)
	router.Get("/catalogo/new", handler.CatalogoNewViewHandler)
	router.Get("/catalogo/detail", handler.CatalogoDetailViewHandler)
	router.Post("/catalogs/create", handler.CreateCatalogHandler)
	router.Post("/catalogs/items/create", handler.CreateCatalogItemHandler)
	router.Post("/catalogs/items/update", handler.UpdateCatalogItemHandler)

	// Handle sprints page view.
	router.Get("/sprints", handler.SprintsViewHandler)
	router.Get("/sprints/detail", handler.TaskDetailViewHandler)

	// Handle API endpoints.
	router.Get("/api/hello-world", handler.ShowContentAPIHandler)
	router.Get("/api/tasks", handler.GetTasksBySprintAPIHandler)
	router.Get("/api/sprints/{id}/summary", handler.GetSprintSummaryAPIHandler)
	router.Get("/api/os-templates/{id}/items", handler.GetOSTemplateItemsAPIHandler)
	router.Post("/projects/create", handler.CreateProjectHandler)
	router.Post("/projects/update", handler.UpdateProjectHandler)
	router.Post("/sprints/create", handler.CreateSprintHandler)
	router.Post("/sprints/update", handler.UpdateSprintHandler)
	router.Post("/tasks/create", handler.CreateTaskHandler)
	router.Post("/tasks/update", handler.UpdateTaskHandler)
	router.Post("/tasks/reorder", handler.ReorderSprintTasksHandler)
	router.Post("/tasks/{id}/os", handler.SaveTaskOSHandler)
	router.Get("/api/tasks/{id}/os", handler.GetOSByTaskAPIHandler)

	// Handle OS routes.
	router.Get("/os/new", handler.NewOSHandler)
	router.Post("/os/create", handler.CreateOSHandler)
	router.Get("/os/templates", handler.OSTemplatesViewHandler)
	router.Get("/os/templates/new", handler.OSTemplateNewViewHandler)
	router.Get("/os/templates/edit", handler.OSTemplateEditViewHandler)
	router.Post("/os/templates/create", handler.CreateOSTemplateHandler)
	router.Post("/os/templates/update", handler.UpdateOSTemplateHandler)
	router.Post("/os/templates/delete", handler.DeleteOSTemplateHandler)

	return router
}
