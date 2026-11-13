package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims mendefinisikan payload data yang disimpan di dalam token JWT
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`

	jwt.RegisteredClaims
}

// GenerateJWT membuat token JWT yang berlaku selama durasi tertentu
func GenerateJWT(userID, email, secretKey string, duration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
