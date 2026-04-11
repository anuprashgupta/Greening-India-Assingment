package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	// UserIDKey is the context key for the authenticated user's ID.
	UserIDKey contextKey = "user_id"
	// UserEmailKey is the context key for the authenticated user's email.
	UserEmailKey contextKey = "user_email"
)

// GetUserID extracts the authenticated user's ID from the request context.
func GetUserID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(UserIDKey).(uuid.UUID)
	return id
}

// GetUserEmail extracts the authenticated user's email from the request context.
func GetUserEmail(ctx context.Context) string {
	email, _ := ctx.Value(UserEmailKey).(string)
	return email
}

// writeError writes a JSON error response without importing the handler package.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Auth returns a middleware that validates JWT tokens from the Authorization header.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			tokenStr := parts[1]

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				slog.Warn("invalid JWT token", "error", err)
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			userIDStr, ok := claims["user_id"].(string)
			if !ok {
				writeError(w, http.StatusUnauthorized, "invalid token: missing user_id")
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid token: malformed user_id")
				return
			}

			email, _ := claims["email"].(string)

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserEmailKey, email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
