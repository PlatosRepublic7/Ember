package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/PlatosRepublic7/ember/internal/auth"
	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/PlatosRepublic7/ember/internal/model_converter"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	DB *database.Queries
}

func NewUserHandler(db *database.Queries) *UserHandler {
	return &UserHandler{DB: db}
}

// Handler for registering a new user
func (h *UserHandler) HandlerCreateUser(c *fiber.Ctx) error {
	// Define the struct matching the expected request payload
	type createUserRequest struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"Error": "Invalid Request Payload",
		})
	}

	// Validate the input
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "Payload is missing required fields",
		})
	}

	// Validate the email address
	ok := auth.IsEmailValid(req.Email)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "The provided email address cannot be validated",
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
		Email:     req.Email,
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

// Handler for getting a user from the database by username, returns only one user
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

// Generate a new access token, or respond with an error
func (h *UserHandler) HandlerRefreshToken(c *fiber.Ctx) error {
	type getRefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token"`
	}
	var req getRefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "Malformed payload",
		})
	}

	accessToken, err := auth.AnalyzeRefreshToken(h.DB, req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": fmt.Sprintf("%v", err),
		})
	}

	if accessToken == "refresh token has expired, login required" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"Token": accessToken,
		})
	} else if accessToken == "refresh token is blacklisted, login required" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"Token": accessToken,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"access": accessToken,
	})
}

// This will generate an access-refresh token pair if successfull
func (h *UserHandler) HandlerLoginUser(c *fiber.Ctx) error {
	type getUserLoginRequest struct {
		Email    string `json:"email"`
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
	user, err := h.DB.GetUserLoginInfo(context.Background(), req.Email)
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

	// We need to check that if there are any refresh tokens in the database for this user,
	// then we need to invalidate them and generate a new one.
	refreshTokenList, err := h.DB.GetAllUserRefreshTokens(context.Background(), user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": fmt.Sprintf("%v", err),
		})
	}

	for i := range refreshTokenList {
		if refreshTokenList[i].IsValid {
			refreshTokenUpdateParams := database.UpdateRefreshTokenParams{
				IsValid:      false,
				UpdatedAt:    time.Now().UTC(),
				RefreshToken: refreshTokenList[i].RefreshToken,
			}

			err := h.DB.UpdateRefreshToken(context.Background(), refreshTokenUpdateParams)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"Error": fmt.Sprintf("%v", err),
				})
			}
		}
	}

	// Generate the access and refresh tokens
	accessTokenString, refreshTokenString, err := auth.GenerateTokenPair(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": fmt.Sprintf("%v", err),
		})
	}

	// Store the newly created Refresh Token in the database
	refreshTokenParams := database.CreateRefreshTokenParams{
		RefreshToken: refreshTokenString,
		IsValid:      true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		UserID:       user.ID,
	}

	dbRefreshToken, err := h.DB.CreateRefreshToken(context.Background(), refreshTokenParams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": "Unable to store refresh token",
		})
	}

	// Return the pair of tokens to the client
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"access":  accessTokenString,
		"refresh": dbRefreshToken.RefreshToken,
	})
}

// Handler for logging out a user. This expects a refresh token, and will invalidate the token,
// preventing any ability to generate new access tokens from it
func (h *UserHandler) HandlerLogoutUser(c *fiber.Ctx) error {
	type updateRefreshToken struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req updateRefreshToken
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "Malformed payload",
		})
	}

	dbRefreshToken, err := h.DB.GetRefreshToken(context.Background(), req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"Error": "Refresh token not found",
		})
	}
	// Construct the Refresh Token parameters for invalidation
	refreshTokenParams := database.UpdateRefreshTokenParams{
		IsValid:      false,
		UpdatedAt:    time.Now().UTC(),
		RefreshToken: dbRefreshToken.RefreshToken,
	}

	err = h.DB.UpdateRefreshToken(context.Background(), refreshTokenParams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": fmt.Sprintf("%v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Success": "Logout complete",
	})
}

// Handler for testing JWT auth middleware
func (h *UserHandler) HandlerAuthTest(c *fiber.Ctx) error {
	// We need to retrieve the user's claims
	userClaims := c.Locals("user")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"Message": "This is a protected endpoint",
		"user":    userClaims,
	})
}
