package handlers

import (
	"context"
	"time"

	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	DB *database.Queries
}

func NewUserHandler(db *database.Queries) *UserHandler {
	return &UserHandler{DB: db}
}

func (h *UserHandler) HandlerCreateUser(c *fiber.Ctx) error {
	// Define the struct matching the expected request payload
	type createUserRequest struct {
		Username string `json:"username"`
	}

	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"Error": "Invalid Request Payload",
		})
	}

	// Validate the input
	if req.Username == "" {
		return c.Status(400).JSON(fiber.Map{
			"Error": "Field 'username' is required",
		})
	}

	// Prepare the parameters for the sqlc generated CreateUser function
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Username:  req.Username,
	}

	user, err := h.DB.CreateUser(context.Background(), params)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"Error": "Failed to create user",
		})
	}

	return c.Status(201).JSON(user)
}
