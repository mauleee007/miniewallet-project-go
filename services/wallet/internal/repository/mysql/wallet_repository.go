package mysql

import (
	"context"

	"wallet-service/internal/domain"

	"gorm.io/gorm"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) domain.WalletRepository {
	return &walletRepository{
		db: db,
	}
}

func (r *walletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	// 1. Ambil koneksi (tx jika ada) lalu simpan wallet
	db := dbFromContext(ctx, r.db)
	return db.Create(wallet).Error
}

func (r *walletRepository) GetByID(ctx context.Context, id string) (*domain.Wallet, error) {
	var wallet domain.Wallet

	// 1. Eksekusi query baca
	db := dbFromContext(ctx, r.db)
	err := db.Where("id = ?", id).First(&wallet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrWalletNotFound
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID string) (*domain.Wallet, error) {
	var wallet domain.Wallet

	// 1. Eksekusi query baca
	db := dbFromContext(ctx, r.db)
	err := db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrWalletNotFound
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) UpdateBalance(ctx context.Context, walletID string, amount int64) error {
	// 1. Ambil koneksi (tx jika ada)
	db := dbFromContext(ctx, r.db)

	// 2. Eksekusi update secara atomik
	result := db.Model(&domain.Wallet{}).
		Where("id = ?", walletID).
		Update("balance", gorm.Expr("balance + ?", amount))

	if result.Error != nil {
		return result.Error
	}

	// 3. Pastikan wallet ditemukan
	if result.RowsAffected == 0 {
		return domain.ErrWalletNotFound
	}

	return nil
}
