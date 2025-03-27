package model_converter

import (
	"database/sql"
	"time"

	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
}

func DatabaseUserToUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
	}
}

type RefreshToken struct {
	ID           int32     `json:"id"`
	RefreshToken string    `json:"refresh_token"`
	IsValid      bool      `json:"is_valid"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func DatabaseTokenToToken(dbToken database.RefreshToken) RefreshToken {
	return RefreshToken{
		ID:           dbToken.ID,
		RefreshToken: dbToken.RefreshToken,
		IsValid:      dbToken.IsValid,
		CreatedAt:    dbToken.CreatedAt,
		UpdatedAt:    dbToken.UpdatedAt,
	}
}

type Message struct {
	ID          uuid.UUID     `json:"id"`
	SenderID    uuid.UUID     `json:"sender_id"`
	RecipientID uuid.UUID     `json:"recipient_id"`
	Content     string        `json:"content"`
	CreatedAt   time.Time     `json:"created_at"`
	ReadAt      sql.NullTime  `json:"read_at"`
	TtlSeconds  sql.NullInt32 `json:"ttl_seconds"`
	ExpiresAt   sql.NullTime  `json:"expires_at"`
	Deleted     bool          `json:"deleted"`
}

func DatabaseMessageToMessage(dbMessage database.Message) Message {
	return Message{
		ID:          dbMessage.ID,
		SenderID:    dbMessage.SenderID,
		RecipientID: dbMessage.RecipientID,
		Content:     dbMessage.Content,
		CreatedAt:   dbMessage.CreatedAt,
		ReadAt:      dbMessage.ReadAt,
		TtlSeconds:  dbMessage.TtlSeconds,
		ExpiresAt:   dbMessage.ExpiresAt,
		Deleted:     dbMessage.Deleted,
	}
}
