package handlers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/PlatosRepublic7/ember/internal/auth"
	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/PlatosRepublic7/ember/internal/model_converter"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var jwtAccessSecret = []byte(os.Getenv("ACCESS_SECRET_KEY"))
var jwtRefreshSecret = []byte(os.Getenv("REFRESH_SECRET_KEY"))

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
		Password string `json:"password"`
	}

	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"Error": "Invalid Request Payload",
		})
	}

	// Validate the input
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "Payload is missing required fields",
		})
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": fmt.Sprintf("%v", err),
		})
	}

	// Prepare the parameters for the sqlc generated CreateUser function
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Username:  req.Username,
		Password:  hashedPassword,
	}

	user, err := h.DB.CreateUser(context.Background(), params)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model_converter.DatabaseUserToUser(user))
}

func (h *UserHandler) HandlerGetUser(c *fiber.Ctx) error {
	type getUserRequest struct {
		Username string `json:"username"`
	}

	var req getUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "Malformed payload",
		})
	}

	user, err := h.DB.GetUserByUsername(context.Background(), req.Username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"GET Error": fmt.Sprintf("%v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(model_converter.DatabaseUserToUser(user))
}

func (h *UserHandler) HandlerLoginUser(c *fiber.Ctx) error {
	type getUserLoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Get Request body
	var req getUserLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "Malformed payload",
		})
	}

	// Retrieve the user from the database
	user, err := h.DB.GetUserLoginInfo(context.Background(), req.Username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"Error": fmt.Sprintf("%v", err),
		})
	}

	// Check the given password against the one in the database
	ok := auth.CheckPasswordHash(req.Password, user.Password)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "Incorrect password",
		})
	}

	// Generate the access token with a short expiration time
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jwtAccessSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": "Could not generate access token",
		})
	}

	// Generate the refresh token with a longer expiration time
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(jwtRefreshSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": "Could not generate refresh token",
		})
	}

	// Return the pair of tokens to the client
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"access":  accessTokenString,
		"refresh": refreshTokenString,
	})
}
