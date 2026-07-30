package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gvillela7/usts/configs"
	"github.com/gvillela7/usts/internal/routes"
	"github.com/gvillela7/usts/web"
)

func main() {
	// Run your server.
	if err := runServer(); err != nil {
		slog.Error("Failed to start server!", "details", err.Error())
		os.Exit(1)
	}
}

// runServer runs a new HTTP server with the loaded environment variables.
func runServer() error {
	// Load configuration.
	config, err := configs.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to database.
	db, err := configs.Connect(config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	slog.Info("Connected to database", "dsn", config.Database.DSN)

	// Run database migration.
	if err := configs.Migrate(db); err != nil {
		return fmt.Errorf("failed to run migration: %w", err)
	}
	slog.Info("Database migration applied successfully")

	// Run database seed.
	if err := configs.Seed(db); err != nil {
		return fmt.Errorf("failed to run seed: %w", err)
	}
	slog.Info("Database seeded successfully")

	router := routes.Setup(db, http.FS(web.StaticFS))

	// Create a new server instance with options from environment variables.
	// For more information, see https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Server.Port),
		Handler:      router, // handle all chi routes
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Send log message.
	slog.Info("Starting server...", "port", config.Server.Port)

	return server.ListenAndServe()
}
