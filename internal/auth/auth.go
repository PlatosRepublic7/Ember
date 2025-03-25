package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/PlatosRepublic7/ember/internal/database"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtAccessSecret = []byte(os.Getenv("ACCESS_SECRET_KEY"))
var jwtRefreshSecret = []byte(os.Getenv("REFRESH_SECRET_KEY"))

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Generate an access and refresh token pair for login functionality, return an error if either cannot be generated
func GenerateTokenPair(user database.GetUserLoginInfoRow) (string, string, error) {

	// Generate the access token with a short expiration time
	accessClaims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jwtAccessSecret)
	if err != nil {
		return "", "", fmt.Errorf("could not generate access token")
	}

	// Generate the refresh token with a longer expiration time
	refreshClaims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(jwtRefreshSecret)
	if err != nil {
		return "", "", fmt.Errorf("could not generate refresh token")
	}

	return accessTokenString, refreshTokenString, nil
}

// Check refreshToken for expiration, if it is valid, generate and return an accessTokenString.
// If it has expired, invalidate it, otherwise return an error
func AnalyzeRefreshToken(DB *database.Queries, refreshToken string) (string, error) {
	// Query the database to check that the given refreshToken exists within our system
	dbRefreshToken, err := DB.GetRefreshToken(context.Background(), refreshToken)
	if err != nil {
		return "", fmt.Errorf("refresh token does not exist")
	}

	token, err := jwt.Parse(dbRefreshToken.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtAccessSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("refresh token cannot be parsed")
	}

	// Need to check if this condition is the only one where jwt.Parse will return non-valid
	if !token.Valid {
		// We need to update our database entry for this refresh token to be invalid
		params := database.UpdateRefreshTokenParams{
			IsValid:      false,
			RefreshToken: dbRefreshToken.RefreshToken,
		}
		err := DB.UpdateRefreshToken(context.Background(), params)
		if err != nil {
			return "", fmt.Errorf("cannot invalidate refresh token")
		}

		return "refresh token has expired, login required", nil
	}
	// Extract the claims and use them to generate a new access token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	accessClaims := jwt.MapClaims{
		"user_id":  claims["user_id"],
		"username": claims["username"],
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jwtAccessSecret)
	if err != nil {
		return "", fmt.Errorf("could not generate access token")
	}
	return accessTokenString, nil
}
