package utils

import (
	"GoWork_7/internal/models"
	"encoding/json"
	"net/http"
)

// ErrorResponse 返回统一的错误响应
func ErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Code:    code,
		Message: message,
	})
}

// SuccessResponse 返回统一的成功响应
func SuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	})
}
