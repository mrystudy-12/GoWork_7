package handlers

import (
	"GoWork_7/internal/models"
	"GoWork_7/internal/service"
	"GoWork_7/internal/utils"
	"encoding/json"
	"net/http"
)

// LoginHandler 登录控制器
type LoginHandler struct {
	loginService *service.LoginService
}

// NewLoginHandler 创建登录控制器实例
func NewLoginHandler(loginService *service.LoginService) *LoginHandler {
	return &LoginHandler{loginService: loginService}
}

// Login 处理用户登录请求
func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	user, token, err := h.loginService.Login(req.Username, req.Password)
	if err != nil {
		utils.AuthLogger.Error("登录失败: %v", err)
		if err.Error() == "ACCOUNT_DISABLED" {
			utils.ErrorResponse(w, http.StatusForbidden, "账户已被禁用")
		} else {
			utils.ErrorResponse(w, http.StatusUnauthorized, "用户名或密码错误")
		}
		return
	}

	utils.SuccessResponse(w, "登录成功", map[string]interface{}{
		"token":    token,
		"id":       user.ID,
		"role":     user.Role,
		"username": user.Username,
	})
}
