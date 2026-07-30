package routes

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	gowebly "github.com/gowebly/helpers"
	handler "github.com/gvillela7/usts/internal/handler"
	"github.com/gvillela7/usts/internal/repository"
)

func Setup(db *sql.DB, static http.FileSystem) *chi.Mux {
	// Create a new chi router.
	router := chi.NewRouter()

	// Use chi middlewares.

	router.Use(middleware.Logger)

	// Initialize repository and handlers.
	repo := repository.NewSQLiteRepository(db)
	handler := handler.New(repo)

	// Handle static files from the embed FS (with a custom handler).
	router.Handle("/static/*", gowebly.StaticFileServerHandler(static))

	// Handle index page view.
	router.Get("/", handler.IndexViewHandler)

	// Handle catalog page view.
	router.Get("/catalogo", handler.CatalogoViewHandler)

	// Handle sprints page view.
	router.Get("/sprints", handler.SprintsViewHandler)
	router.Get("/sprints/detail", handler.TaskDetailViewHandler)

	// Handle API endpoints.
	router.Get("/api/hello-world", handler.ShowContentAPIHandler)
	router.Get("/api/tasks", handler.GetTasksBySprintAPIHandler)
	router.Post("/projects/create", handler.CreateProjectHandler)
	router.Post("/projects/update", handler.UpdateProjectHandler)
	router.Post("/sprints/create", handler.CreateSprintHandler)
	router.Post("/sprints/update", handler.UpdateSprintHandler)
	router.Post("/tasks/create", handler.CreateTaskHandler)

	// Handle OS routes.
	router.Get("/os/new", handler.NewOSHandler)
	router.Post("/os/create", handler.CreateOSHandler)

	return router
}
