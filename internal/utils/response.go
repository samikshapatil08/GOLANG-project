package utils

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

func SendJSON(writer http.ResponseWriter, statusCode int, payload interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	response := SuccessResponse{
		Data: payload,
	}
	_ = json.NewEncoder(writer).Encode(response)
}

func SendError(writer http.ResponseWriter, statusCode int, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	response := ErrorResponse{
		Error: message,
	}
	_ = json.NewEncoder(writer).Encode(response)
}
