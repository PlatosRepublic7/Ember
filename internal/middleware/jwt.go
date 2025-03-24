package middleware

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var jwtAccessSecret = []byte("the_access_secret")

// JWTAuthMiddleware validates the access token
func JWTAuthMiddleware(c *fiber.Ctx) error {
	// Get the token from the Authorization Header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"Error": "Missing Access Token",
		})
	}

	// Identify and remove the "Bearer" prefix
	authVals := strings.Split(authHeader, " ")
	if len(authVals) < 2 {
		return errors.New("malformed authorization header")
	}

	if authVals[0] != "Bearer" {
		return errors.New("incorrect authorization content, expected 'Bearer'")
	}

	tokenString := authVals[1]

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtAccessSecret, nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"Error": "Invalid or expired access token",
		})
	}

	// Extract the claims and add them to the context
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"Error": "Invalid token claims",
		})
	}

	c.Locals("user", claims)
	return c.Next()
}
