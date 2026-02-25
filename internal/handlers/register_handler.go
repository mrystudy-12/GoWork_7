package handlers

import (
	"GoWork_7/internal/models"
	"GoWork_7/internal/service"
	"GoWork_7/internal/utils"
	"encoding/json"
	"net/http"
)

// RegisterHandler 注册控制器
type RegisterHandler struct {
	registerService *service.RegisterService
}

// NewRegisterHandler 创建注册控制器实例
func NewRegisterHandler(registerService *service.RegisterService) *RegisterHandler {
	return &RegisterHandler{registerService: registerService}
}

// Register 处理用户注册请求
func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// 解析 JSON 请求体
	var req models.LoginRequest // 复用 LoginRequest 结构，因为字段相同
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "无效的 JSON 数据")
		return
	}

	if req.Username == "" || len(req.Password) != 6 {
		utils.ErrorResponse(w, http.StatusBadRequest, "格式错误：用户名不能为空且密码必须为6位")
		return
	}

	uid, err := h.registerService.Register(req.Username, req.Password)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "注册失败：用户名可能已被占用")
		return
	}

	// 注册成功后直接生成 token
	token, _ := utils.GenerateToken(uid, req.Username, "common")

	utils.SuccessResponse(w, "注册成功", map[string]interface{}{
		"user_id": uid,
		"token":   token,
	})
}
