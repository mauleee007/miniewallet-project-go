package response

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Status  string      `json: "status"`
	Message string      `json: "message"`
	Data    interface{} `json: "data,amitempty"`
}

type ErrorResponse struct {
	Status  string `json: "status"`
	Message string `json: "message"`
}

func JSON(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	//1. Set Content-type header JSON
	w.Header().Set("Content-Type", "application/json")

	//2. Set HTTP status Code
	w.WriteHeader(statusCode)

	//3. Susun dan encode response struct
	json.NewEncoder(w).Encode(SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, statusCode int, message string) {
	//1. Set Content-type header JSON
	w.Header().Set("Content-Type", "application/json")

	//2. Set HTTP Status Code
	w.WriteHeader(statusCode)

	//3. Susun dan encode error struct
	json.NewEncoder(w).Encode(ErrorResponse{
		Status:  "error",
		Message: message,
	})
}
