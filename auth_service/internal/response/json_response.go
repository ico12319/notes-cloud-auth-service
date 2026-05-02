package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SuccessResponse[T any] struct {
	Data T `json:"data"`
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

func WriteError(w http.ResponseWriter, statusCode int, code string, message string) {
	WriteJSON(w, statusCode, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func WriteSuccess[T any](w http.ResponseWriter, statusCode int, data T) {
	WriteJSON(w, statusCode, SuccessResponse[T]{
		Data: data,
	})
}
