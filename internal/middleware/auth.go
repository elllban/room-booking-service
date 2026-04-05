package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(jwtSecret []byte) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendAuthError(w, "UNAUTHORIZED", "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			sendAuthError(w, "UNAUTHORIZED", "invalid authorization header format")
			return
		}

		tokenString := parts[1]

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			sendAuthError(w, "UNAUTHORIZED", "invalid or expired token")
			return
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			sendAuthError(w, "UNAUTHORIZED", "invalid token claims")
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			sendAuthError(w, "UNAUTHORIZED", "invalid user id in token")
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			sendAuthError(w, "UNAUTHORIZED", "invalid role in token")
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "role", role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sendAuthError(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":{"code":"` + code + `","message":"` + message + `"}}`))
}
