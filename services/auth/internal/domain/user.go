package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Username     string    `gorm:"unique;type:varchar(50);not null" json:"username"`
	Email        string    `gorm:"unique;type:varchar(100);not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"` // tanda "-" agar tidak terekspos di JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

type AuthUsecase interface {
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
}

// RegisterRequest menampung payload JSON saat user mendaftar
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest menampung payload JSON saat user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse adalah data user yang aman untuk dikembalikan ke client (tanpa hash password)
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponse adalah payload balikan setelah registrasi atau login berhasil
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
