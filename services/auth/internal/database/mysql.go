package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitMySQL menghubungkan aplikasi ke MySQL menggunakan GORM dan mengonfigurasi connection pool
func InitMySQL(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Menampilkan query SQL di terminal saat debug
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi MySQL: %w", err)
	}

	// Konfigurasi connection pooling melalui underlying database sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan instance generic sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)           // Koneksi idle maksimum di pool
	sqlDB.SetMaxOpenConns(100)          // Koneksi terbuka maksimum ke DB
	sqlDB.SetConnMaxLifetime(time.Hour) // Batas umur koneksi dapat digunakan kembali

	return db, nil
}
