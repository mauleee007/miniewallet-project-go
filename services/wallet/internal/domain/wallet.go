package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrWalletNotFound = errors.New("wallet not found")
	ErrInvalidAmount  = errors.New("invalid amount")
)

type Wallet struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID    string    `gorm:"uniqueIndex;type:varchar(36);not null" json:"user_id"`
	Balance   int64     `gorm:"not null;default:0" json:"balance"` // Menggunakan int64 untuk presisi keuangan
	Currency  string    `gorm:"type:varchar(3);default:'IDR'" json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TopUpRequest struct {
	Amount int64 `json:"amount"`
}

type WalletResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WalletRepository interface {
	Create(ctx context.Context, wallet *Wallet) error
	GetByID(ctx context.Context, id string) (*Wallet, error)
	GetByUserID(ctx context.Context, userID string) (*Wallet, error)
	UpdateBalance(ctx context.Context, walletID string, amount int64) error
}

type WalletUsecase interface {
	GetBalance(ctx context.Context, userID string) (*WalletResponse, error)
	TopUp(ctx context.Context, userID string, req *TopUpRequest) (*TransactionResponse, error)
}
