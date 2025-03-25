package model_converter

import (
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
	Password  string    `json:"password"`
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
