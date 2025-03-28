package handlers

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/PlatosRepublic7/ember/internal/auth"
	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/PlatosRepublic7/ember/internal/model_converter"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MessageHandler struct {
	DB *database.Queries
}

func NewMessageHandler(db *database.Queries) *MessageHandler {
	return &MessageHandler{DB: db}
}

// This handler will create a new message in the database (we will eventually add encrypion logic here
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

	userID, err := auth.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": fmt.Sprintf("cannot access claims: %v", err),
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

// Handler for getting all messages sent/received to/from a specific user
func (h *MessageHandler) HandlerGetMessages(c *fiber.Ctx) error {
	// The request will have a query-string parameter: ?type=sent or ?type=received
	// If empty/missing, this will return a list of all sent and received messages associated with the user
	reqType := c.Query("type", "")

	// ?username=SomeUsername will return messages with that specific user
	reqUsername := c.Query("username", "")

	var isUsername bool
	var qUserID uuid.UUID

	if reqUsername == "" {
		isUsername = false
	} else {
		// username query-tag is included and non-empty. We need to get their id from the database
		qUser, err := h.DB.GetUserByUsername(c.UserContext(), reqUsername)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"Error": fmt.Sprintf("Requested user '%s' does not exist. %v", reqUsername, err),
			})
		}

		qUserID = qUser.ID
		isUsername = true
	}

	// We first need to get the requesting user's id from the token
	userID, err := auth.GetUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"Error": fmt.Sprintf("cannot access claims: %v", err),
		})
	}

	var messages []database.Message

	switch reqType {
	case "sent":
		if isUsername {
			getSentMessagesToNamedUserParams := database.GetSentMessagesToNamedUserParams{
				SenderID:    userID,
				RecipientID: qUserID,
			}

			messages, err = h.DB.GetSentMessagesToNamedUser(c.UserContext(), getSentMessagesToNamedUserParams)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"Error": fmt.Sprintf("%v", err),
				})
			}
		} else {
			// Query the database for all messages sent by this user (that have not been deleted)
			messages, err = h.DB.GetSentMessagesFromThisUser(c.UserContext(), userID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"Error": fmt.Sprintf("%v", err),
				})
			}
		}

	case "received":
		if isUsername {
			getReceivedMessagesFromNamedUserParams := database.GetReceivedMessagesFromNamedUserParams{
				RecipientID: userID,
				SenderID:    qUserID,
			}

			messages, err = h.DB.GetReceivedMessagesFromNamedUser(c.UserContext(), getReceivedMessagesFromNamedUserParams)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"Error": fmt.Sprintf("%v", err),
				})
			}
		} else {
			messages, err = h.DB.GetReceivedMessagesToThisUser(c.UserContext(), userID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"Error": fmt.Sprintf("%v", err),
				})
			}
		}

	default:
		if isUsername {
			getMessageHistoryWithNamedUserParams := database.GetMessageHistoryWithNamedUserParams{
				SenderID:    userID,
				RecipientID: qUserID,
			}

			messages, err = h.DB.GetMessageHistoryWithNamedUser(c.UserContext(), getMessageHistoryWithNamedUserParams)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"Error": fmt.Sprintf("%v", err),
				})
			}
		} else {
			messages, err = h.DB.GetUserMessageHistory(c.UserContext(), userID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"Error": fmt.Sprintf("%v", err),
				})
			}
		}
	}

	// Convert the database.Message model fields to snake_case json fields
	convertedMessages := make([]model_converter.Message, len(messages))
	for i := range messages {
		convMessage := model_converter.DatabaseMessageToMessage(messages[i])
		convertedMessages[i] = convMessage
	}

	return c.Status(fiber.StatusOK).JSON(convertedMessages)
}
