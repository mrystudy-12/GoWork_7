package models

// User 用户模型结构体
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"` // 关键：转 JSON 时隐藏密码
	LastLogin string `json:"last_login"`
	Role      string `json:"role"`
	Enable    bool   `json:"enable"`
	Avatar    string `json:"avatar,omitempty"`
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest 注册请求结构体
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Response 通用响应结构体
type Response struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// APIResponse API 响应结构体
type APIResponse struct {
	Success bool        `json:"success"` // 必须添加，它是前后端的“生死线”
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
