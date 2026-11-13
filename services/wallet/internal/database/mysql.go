package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitMySQL(dsn string) (*gorm.DB, error) {
	// 1. Buka koneksi GORM ke database MySQL
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database MySQL wallet_db: %w", err)
	}

	// 2. Ambil generic instance sql.DB untuk mengonfigurasi pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil instance generic sql.DB: %w", err)
	}

	// 3. Konfigurasi Connection Pooling
	sqlDB.SetMaxIdleConns(10)           // Maksimal 10 koneksi idle
	sqlDB.SetMaxOpenConns(100)          // Maksimal 100 koneksi terbuka simultan
	sqlDB.SetConnMaxLifetime(time.Hour) // Daur ulang koneksi setelah 1 jam

	return db, nil
}
