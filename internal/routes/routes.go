package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/PlatosRepublic7/ember/internal/handlers"
)

func SetupRoutes(app *fiber.App, dbInstance *database.Queries) {
	app.Get("/healthc", handlers.HealthCheck)

	userHandler := handlers.NewUserHandler(dbInstance)

	// Create URI group for app
	v1 := app.Group("/v1")
	v1.Post("/users", userHandler.HandlerCreateUser)
}
