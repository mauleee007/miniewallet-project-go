package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengenkripsi plain text password menggunakan Bcrypt dengan cost minimal 10
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash membandingkan plain text password dengan hash yang tersimpan di DB
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
