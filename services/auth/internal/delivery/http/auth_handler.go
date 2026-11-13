package http

import (
	"auth-service/internal/domain"
	"auth-service/pkg/response"
	"encoding/json"

	"net/http"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

// NewAuthHandler menginisiasi handler hhtp untuk modul authentication
func NewAuthHandler(authUseCase domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUseCase,
	}
}

// RegisterRoutes mendaftarkan endpoint rute auth ke Chi Router
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
}

// Register menangani HTTP POST pendaftaran user baru
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid payload format")
		return
	}
	res, err := h.authUsecase.Register(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, "Register Successfully", res)
}

// Login menangani HTTP POST Authentication user
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid payload format")
		return
	}

	res, err := h.authUsecase.Login(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Login Successfully", res)
}
