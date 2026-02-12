package utils

import "net/http"

// SetCORSHeaders 设置跨域请求头
// methods: 允许的HTTP方法，如 "GET, POST, PUT, DELETE, OPTIONS"
func SetCORSHeaders(w http.ResponseWriter, methods string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "New-Token")
}

// HandleOPTIONS 处理OPTIONS预检请求
func HandleOPTIONS(w http.ResponseWriter) {
	SetCORSHeaders(w, "GET, POST, PUT, DELETE, OPTIONS")
	w.WriteHeader(http.StatusOK)
}
