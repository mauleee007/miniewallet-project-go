package middleware

import (
	"context"
	"net/http"
	"strings"
	"wallet-service/pkg/response"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//1. Ekstrak header authorization
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer") {
				response.Error(w, http.StatusUnauthorized, "Missing or invalid token format")
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer")

			//2. Parse & Verifikasi token
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				response.Error(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			//3. Ekstrak klaim user_id
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.Error(w, http.StatusUnauthorized, "Invalid token claims")
				return
			}

			userID, ok := claims["user_id"].(string)
			if !ok || userID == "" {
				response.Error(w, http.StatusUnauthorized, "user_id not found in token")
				return
			}

			//4. lanjutkan ke handler dengan context baru
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
