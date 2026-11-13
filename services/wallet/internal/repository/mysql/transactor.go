package mysql

import (
	"context"

	"wallet-service/internal/domain"

	"gorm.io/gorm"
)

type txKey struct{}

type transactor struct {
	db *gorm.DB
}

func NewTransactor(db *gorm.DB) domain.Transactor {
	return &transactor{db: db}
}

func (t *transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 1. Buka transaksi, sisipkan tx ke context, jalankan fn
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

func dbFromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	// 1. Gunakan tx jika ada di context
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	// 2. Fallback ke koneksi biasa
	return fallback.WithContext(ctx)
}
