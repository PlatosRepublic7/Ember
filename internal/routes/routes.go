package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/PlatosRepublic7/ember/internal/handlers"
	"github.com/PlatosRepublic7/ember/internal/middleware"
)

func SetupRoutes(app *fiber.App, dbInstance *database.Queries) {
	app.Get("/healthc", handlers.HealthCheck)

	// Create URI group for app
	v1 := app.Group("/v1/auth")

	// Create a userHandler
	userHandler := handlers.NewUserHandler(dbInstance)

	// All non-protected endpoints
	v1.Post("/register", userHandler.HandlerCreateUser)
	v1.Get("/users", userHandler.HandlerGetUser)
	v1.Post("/login", userHandler.HandlerLoginUser)
	v1.Post("/logout", userHandler.HandlerLogoutUser)
	v1.Post("/refresh", userHandler.HandlerRefreshToken)

	// Group for all auth protected endpoints
	protected := app.Group("/v1", middleware.JWTAuthMiddleware)
	protected.Get("/test", userHandler.HandlerAuthTest)
}
