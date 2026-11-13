package main

import (
	"log"
	"net/http"

	"wallet-service/internal/config"
	"wallet-service/internal/database"
	walletHttp "wallet-service/internal/delivery/http"
	"wallet-service/internal/domain"
	"wallet-service/internal/repository/mysql"
	"wallet-service/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 1. Muat konfigurasi lingkungan
	cfg := config.LoadConfig()

	// 2. Hubungkan ke MySQL
	db, err := database.InitMySQL(cfg.DB_DSN)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	// 3. Auto-Migrate tabel (Wallet & Transaction)
	err = db.AutoMigrate(&domain.Wallet{}, &domain.Transaction{})
	if err != nil {
		log.Fatalf("Gagal auto-migrate: %v", err)
	}

	// 4. Inisialisasi Repositori
	walletRepo := mysql.NewWalletRepository(db)
	transactionRepo := mysql.NewTransactionRepository(db)

	// 5. Inisialisasi Usecase
	transactor := mysql.NewTransactor(db)
	walletUC := usecase.NewWalletUsecase(walletRepo, transactionRepo, transactor)

	// 6. Setup Router utama
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 7. Daftarkan Handler
	walletHttp.NewWalletHandler(r, walletUC, cfg.JWTSecret)

	// 8. Nyalakan Server
	log.Printf("Wallet Service (Modul 2) berjalan di port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server berhenti: %v", err)
	}
}
