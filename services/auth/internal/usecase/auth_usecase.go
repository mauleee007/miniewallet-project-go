package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"auth-service/internal/domain"
	"auth-service/pkg/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type authUseCase struct {
	userRepo    domain.UserRepository
	redisClient *redis.Client
	jwtSecret   string
}

// NewAuthUseCase menginisiasi usecase auth dengan depedency injection
func NewAuthUseCase(userRepo domain.UserRepository, redisClient *redis.Client, jwtSecret string) domain.AuthUsecase {
	return &authUseCase{
		userRepo:    userRepo,
		redisClient: redisClient,
		jwtSecret:   jwtSecret,
	}
}

// Register memvalidasi input, mengecek email duplikat, membuat user baru, dan generate JWT.
func (u *authUseCase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	//1. Validate input
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, errors.New("username, email, and password must be fill")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	//2. Cek apakah Email sudah terdaftar
	existingUser, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("Email already used")
	}

	//3. Hash Password dengan bcrypt
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("Failed to encryp password")
	}

	//4. Membuat entity User baru dengan UUID
	newUser := &domain.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	//5. Simpan ke database MYSQL
	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, errors.New("Failed to create account")
	}

	//6. Generate Token JWT : Aktif selama 24 jam
	token, err := utils.GenerateJWT(newUser.ID, newUser.Email, u.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, errors.New("Failed to generate authentication token")
	}

	//7. Simpan sesi aktif ke redis (cache)
	sessionKey := "session:user:" + newUser.ID
	_ = u.redisClient.Set(ctx, sessionKey, token, 24*time.Hour).Err()

	return &domain.AuthResponse{
		Token: token,
		User: domain.UserResponse{
			ID:        newUser.ID,
			Username:  newUser.Username,
			Email:     newUser.Email,
			CreatedAt: newUser.CreatedAt,
		},
	}, nil

}

// Login implements [domain.AuthUsecase].
func (u *authUseCase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	//1. Validasi input
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, errors.New("Email and password must be fill")
	}

	//2. Cari data user berdasarkan email
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		//Pesan error sengaja dibuat untuk keamanan
		return nil, errors.New("Email and password combination are worng")
	}
	//3. Verifikasi kesesuaian hash password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("Email and password combination are wrong")
	}
	// 4. Generate token JWT(aktif 24 jam)
	token, err := utils.GenerateJWT(user.ID, user.Email, u.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, errors.New("Failed to generate authentication token")
	}
	//5. Perbarui sesi di Redis
	sessionKey := "session:user:" + user.ID
	_ = u.redisClient.Set(ctx, sessionKey, token, 24*time.Hour).Err()

	return &domain.AuthResponse{
		Token: token,
		User: domain.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}
