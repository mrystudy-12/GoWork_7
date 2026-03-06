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
