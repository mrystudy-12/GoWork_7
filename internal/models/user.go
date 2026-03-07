package models

import "time"

// User 基础用户模型
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"` // 敏感信息，返回时忽略
	Role      string    `json:"role"`
	LastLogin time.Time `json:"last_login"`
	Enable    bool      `json:"enable"`
	Avatar    string    `json:"avatar"`
}

// LoginResponse 登录成功后的响应数据结构
type LoginResponse struct {
	Token    string `json:"token"`
	ID       int64  `json:"id"`
	Role     string `json:"role"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
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

// APIResponse API 响应结构体
type APIResponse struct {
	Success bool        `json:"success"` // 必须添加，它是前后端的“生死线”
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
