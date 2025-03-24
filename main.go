package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/PlatosRepublic7/ember/internal/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	// Get environment variables and initialize port number
	godotenv.Load()

	portString := os.Getenv("SERVER_PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not found in the environment")
	}

	// Connect to database
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Cannot connect to database")
	}

	apiCfg := apiConfig{
		DB: database.New(conn),
	}

	log.Println("Config:", apiCfg)

	// Create the Fiber application and initialize logger, recovery, and cors
	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	fmt.Println("Server running on port", portString)
	portString = ":" + portString

	routes.SetupRoutes(app, apiCfg.DB)
	app.Listen(portString)
}
