package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWT creates a token valid for 30 days
func GenerateJWT(userID, email string) (string, error) {
	secretString := os.Getenv("JWT_SECRET")
	if secretString == "" {
		secretString = "fallback-secret-do-not-use-in-production"
	}

	jwtSecretKey := []byte(secretString)

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}
