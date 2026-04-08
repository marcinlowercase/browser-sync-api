package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// RequireAuth wraps handlers to ensure they have a valid JWT
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized - Missing Token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			var jwtSecretString = os.Getenv("JWT_SECRET")
			if jwtSecretString == "" {
				jwtSecretString = "fallback-secret-do-not-use-in-production"
			}
			return []byte(jwtSecretString), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized - Invalid Token", http.StatusUnauthorized)
			return
		}

		// Extract user_id and put it in the Request Context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized - Invalid Claims", http.StatusUnauthorized)
			return
		}

		userID := claims["user_id"].(string)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		// Pass the request to the next handler, now loaded with the user_id
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
