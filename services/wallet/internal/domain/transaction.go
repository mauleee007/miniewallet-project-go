package domain

import (
	"context"
	"time"
)

type TransactionType string

const (
	TxTypeTopUp    TransactionType = "TOPUP"
	TxTypeTransfer TransactionType = "TRANSFER"
)

type TransactionStatus string

const (
	TxStatusSuccess TransactionStatus = "SUCCESS"
	TxStatusFailed  TransactionStatus = "FAILED"
)

type Transaction struct {
	ID                string            `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ReferenceID       string            `gorm:"uniqueIndex;type:varchar(100);not null" json:"reference_id"`
	SenderWalletID    *string           `gorm:"type:varchar(36);index" json:"sender_wallet_id,omitempty"`
	RecipientWalletID string            `gorm:"type:varchar(36);index;not null" json:"recipient_wallet_id"`
	Amount            int64             `gorm:"not null" json:"amount"`
	Type              TransactionType   `gorm:"type:varchar(20);not null" json:"type"`
	Status            TransactionStatus `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt         time.Time         `json:"created_at"`
}

type TransactionResponse struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id,omitempty"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	Status      string    `json:"status"`
	ReferenceID string    `json:"reference_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByWalletID(ctx context.Context, walletID string, limit, offset int) ([]Transaction, error)
}
