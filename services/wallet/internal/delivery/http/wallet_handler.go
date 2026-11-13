package http

import (
	"encoding/json"
	"net/http"
	"wallet-service/internal/domain"
	"wallet-service/pkg/middleware"
	"wallet-service/pkg/response"

	"github.com/go-chi/chi/v5"
)

type walletHandler struct {
	usecase domain.WalletUsecase
}

func NewWalletHandler(r *chi.Mux, u domain.WalletUsecase, jwtSecret string) {
	handler := &walletHandler{
		usecase: u,
	}
	// 1. Daftarkan rute dengan middleware pelindung
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(jwtSecret))
		r.Get("/api/v1/wallets/balance", handler.GetBalance)
		r.Post("/api/v1/wallets/topup", handler.TopUp)
	})
}

func (h *walletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	// 1. Ekstrak userID dari context middleware
	userID := r.Context().Value(middleware.UserIDKey).(string)
	// 2. Eksekusi Usecase
	res, err := h.usecase.GetBalance(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 3. Balikan respons sukses
	response.JSON(w, http.StatusOK, "Berhasil mendapatkan saldo", res)
}

func (h *walletHandler) TopUp(w http.ResponseWriter, r *http.Request) {
	// 1. Ekstrak userID dari context middleware
	userID := r.Context().Value(middleware.UserIDKey).(string)
	// 2. Parse body request
	var req domain.TopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}
	// 3. Eksekusi Usecase
	res, err := h.usecase.TopUp(r.Context(), userID, &req)
	if err != nil {
		if err == domain.ErrInvalidAmount {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 4. Balikan respons sukses
	response.JSON(w, http.StatusOK, "Top-up berhasil diproses", res)
}
