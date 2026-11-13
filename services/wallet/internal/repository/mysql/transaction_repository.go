package mysql

import (
	"context"
	"wallet-service/internal/domain"

	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

// Create implements [domain.TransactionRepository].
func (r *transactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	// 1. Ambil koneksi (tx jika ada) lalu catat ledger
	db := dbFromContext(ctx, r.db)
	return db.Create(tx).Error
}

// GetByWalletID implements [domain.TransactionRepository].
func (r *transactionRepository) GetByWalletID(ctx context.Context, walletID string, limit int, offset int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	//1. eksekusi query dengan limit/offset
	err := r.db.WithContext(ctx).Where("wallet_id = ?", walletID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func NewTransactionRepository(db *gorm.DB) domain.TransactionRepository {
	return &transactionRepository{
		db: db,
	}
}
