package handlers

import (
	"GoWork_7/internal/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 测试CORS头设置
func TestCORSHeaders(t *testing.T) {
	// 创建一个测试响应记录器
	w := httptest.NewRecorder()

	// 调用SetCORSHeaders函数
	utils.SetCORSHeaders(w, "GET, POST, PUT, DELETE, OPTIONS")

	// 检查CORS头是否正确设置
	headers := w.Header()
	if headers.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin to be *, got %s", headers.Get("Access-Control-Allow-Origin"))
	}

	if headers.Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Expected Access-Control-Allow-Methods to be GET, POST, PUT, DELETE, OPTIONS, got %s", headers.Get("Access-Control-Allow-Methods"))
	}

	if headers.Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Errorf("Expected Access-Control-Allow-Headers to be Content-Type, Authorization, got %s", headers.Get("Access-Control-Allow-Headers"))
	}

	if headers.Get("Access-Control-Expose-Headers") != "New-Token" {
		t.Errorf("Expected Access-Control-Expose-Headers to be New-Token, got %s", headers.Get("Access-Control-Expose-Headers"))
	}
}

// 测试错误响应
func TestErrorResponse(t *testing.T) {
	// 创建一个测试响应记录器
	w := httptest.NewRecorder()

	// 调用ErrorResponse函数
	utils.ErrorResponse(w, http.StatusBadRequest, "Test error message")

	// 检查响应状态码
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	// 检查响应内容
	var response struct {
		Success bool   `json:"success"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success != false {
		t.Errorf("Expected success to be false, got %v", response.Success)
	}

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected code to be %d, got %d", http.StatusBadRequest, response.Code)
	}

	if response.Message != "Test error message" {
		t.Errorf("Expected message to be 'Test error message', got '%s'", response.Message)
	}
}

// 测试成功响应
func TestSuccessResponse(t *testing.T) {
	// 创建一个测试响应记录器
	w := httptest.NewRecorder()

	// 准备测试数据
	testData := map[string]string{
		"key": "value",
	}

	// 调用SuccessResponse函数
	utils.SuccessResponse(w, "Test success message", testData)

	// 检查响应状态码
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// 检查响应内容
	var response struct {
		Success bool        `json:"success"`
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success != true {
		t.Errorf("Expected success to be true, got %v", response.Success)
	}

	if response.Code != http.StatusOK {
		t.Errorf("Expected code to be %d, got %d", http.StatusOK, response.Code)
	}

	if response.Message != "Test success message" {
		t.Errorf("Expected message to be 'Test success message', got '%s'", response.Message)
	}

	// 检查数据是否正确
	dataMap, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data to be a map[string]interface{}")
	}

	if dataMap["key"] != "value" {
		t.Errorf("Expected data.key to be 'value', got '%v'", dataMap["key"])
	}
}

// 测试OPTIONS请求处理
func TestHandleOPTIONS(t *testing.T) {
	// 创建一个测试响应记录器
	w := httptest.NewRecorder()

	// 调用HandleOPTIONS函数
	utils.HandleOPTIONS(w)

	// 检查响应状态码
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// 检查CORS头是否正确设置
	headers := w.Header()
	if headers.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin to be *, got %s", headers.Get("Access-Control-Allow-Origin"))
	}
}

// 测试文件类型验证
func TestIsAllowedType(t *testing.T) {
	// 测试允许的文件类型
	allowedTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
	}

	for _, contentType := range allowedTypes {
		if !isAllowedType(contentType) {
			t.Errorf("Expected %s to be allowed", contentType)
		}
	}

	// 测试不允许的文件类型
	disallowedTypes := []string{
		"image/bmp",
		"text/plain",
		"application/json",
	}

	for _, contentType := range disallowedTypes {
		if isAllowedType(contentType) {
			t.Errorf("Expected %s to be disallowed", contentType)
		}
	}
}

// 测试登录请求 - 跳过，因为需要数据库连接
func TestLogin(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}

// 测试注册请求 - 跳过，因为需要数据库连接
func TestRegister(t *testing.T) {
	t.Skip("Skipping test that requires database connection")
}
