package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/PlatosRepublic7/ember/internal/handlers"
)

func SetupRoutes(app *fiber.App, dbInstance *database.Queries) {
	app.Get("/healthc", handlers.HealthCheck)

	// Create URI group for app
	v1 := app.Group("/v1")

	// Create a userHandler
	userHandler := handlers.NewUserHandler(dbInstance)
	v1.Post("/register", userHandler.HandlerCreateUser)
	v1.Get("/users", userHandler.HandlerGetUser)
	v1.Post("/login", userHandler.HandlerLoginUser)
}
