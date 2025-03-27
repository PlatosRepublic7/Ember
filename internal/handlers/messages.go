package handlers

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/PlatosRepublic7/ember/internal/model_converter"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type MessageHandler struct {
	DB *database.Queries
}

func NewMessageHandler(db *database.Queries) *MessageHandler {
	return &MessageHandler{DB: db}
}

// This handler will create a new message in the database (we will eventually add encrypion logging here
// or in some middleware)
func (h *MessageHandler) HandlerCreateMessage(c *fiber.Ctx) error {
	// Define the expected request payload as a struct
	type createMessageRequest struct {
		Username   string `json:"username"`
		Content    string `json:"content"`
		TtlSeconds *int32 `json:"ttl_seconds"`
	}

	var req createMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"Error": "Malformed payload",
		})
	}

	// Search the database for the recipient's id
	rUser, err := h.DB.GetUserByUsername(c.UserContext(), req.Username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"Error": "Requested recipient does not exist",
		})
	}

	// Retrieve the requesting users Token
	claims := c.Locals("user").(jwt.MapClaims)

	// Get the user_id from the claims
	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": fmt.Sprintf("cannot access claims %v", err),
		})
	}

	// We need to check whether TtlSeconds is null (nil) or present
	var ttlSeconds sql.NullInt32
	if req.TtlSeconds == nil {
		ttlSeconds = sql.NullInt32{
			Valid: false,
		}
	} else {
		ttlSeconds = sql.NullInt32{
			Int32: *req.TtlSeconds,
			Valid: true,
		}
	}

	// And now calculate the expiration time (if needed)
	var expiresAt sql.NullTime
	if ttlSeconds.Valid {
		expirationTime := time.Now().Add(time.Duration(ttlSeconds.Int32) * time.Second).UTC()
		expiresAt = sql.NullTime{
			Time:  expirationTime,
			Valid: true,
		}
	} else {
		expiresAt = sql.NullTime{
			Valid: false,
		}
	}

	// Create new message
	createMessageParams := database.CreateMessageParams{
		ID:          uuid.New(),
		SenderID:    userID,
		RecipientID: rUser.ID,
		Content:     req.Content,
		CreatedAt:   time.Now().UTC(),
		TtlSeconds:  ttlSeconds,
		ExpiresAt:   expiresAt,
	}

	message, err := h.DB.CreateMessage(c.UserContext(), createMessageParams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": fmt.Sprintf("%v", err),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model_converter.DatabaseMessageToMessage(message))
}
