package main

import (
	"fmt"
	"log"
	"net/http"

	"auth-service/internal/config"
	"auth-service/internal/database"
	httpDelivery "auth-service/internal/delivery/http"
	"auth-service/internal/domain"
	repoMySQL "auth-service/internal/repository/mysql"
	"auth-service/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 1. Muat Konfigurasi dari Environment
	cfg := config.LoadConfig()

	// 2. Hubungkan ke MySQL
	db, err := database.InitMySQL(cfg.DB_DSN)
	if err != nil {
		log.Fatalf("Inisialisasi MySQL gagal: %v", err)
	}

	// 3. Jalankan Auto-Migrate untuk tabel domain
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		log.Fatalf("Auto-migrate database gagal: %v", err)
	}
	log.Println("Migrasi tabel users berhasil!")

	// 4. Hubungkan ke Redis
	rdb, err := database.InitRedis(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Inisialisasi Redis gagal: %v", err)
	}
	log.Println("Koneksi Redis berhasil!")

	// 5. Inisialisasi Dependency (Clean Architecture Wiring)
	userRepo := repoMySQL.NewUserRepository(db)
	authUsecase := usecase.NewAuthUseCase(userRepo, rdb, cfg.JWTSecret)
	authHandler := httpDelivery.NewAuthHandler(authUsecase)

	// 6. Siapkan Router Chi & Global Middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)    // Log setiap HTTP request yang masuk
	r.Use(middleware.Recoverer) // Menangkap panic agar server tidak crash

	// Health Check Route
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("auth-service is healthy"))
	})

	// Mount Group Route Auth API
	r.Route("/api/v1/auth", func(r chi.Router) {
		authHandler.RegisterRoutes(r)
	})

	// 7. Jalankan HTTP Server
	fmt.Printf("🚀 Auth Service berjalan di port :%s...\n", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
