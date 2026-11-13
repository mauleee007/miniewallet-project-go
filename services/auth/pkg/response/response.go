package response

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// JSON mengirim response berformat JSON untuk kasus berhasil (200, 201, dll)
func JSON(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(resp)
}

// Error mengirim response berformat JSON untuk kasus gagal (400, 401, 500, dll)
func Error(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Status:  "error",
		Message: message,
	}

	json.NewEncoder(w).Encode(resp)
}
