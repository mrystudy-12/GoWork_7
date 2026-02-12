package handlers

import (
	"GoWork_7/internal/database"
	"GoWork_7/internal/models"
	"GoWork_7/internal/utils"
	"encoding/json"
	"net/http"
)

// Login  处理用户登录请求
// 接收JSON格式的用户名和密码，验证后返回包含token的响应
func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		utils.AuthLogger.Info("登录请求方法错误: %s", r.Method)
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.AuthLogger.Error("解析登录请求失败: %v", err)
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	user, err := checkLogin(req.Username, req.Password)
	if err != nil {
		utils.AuthLogger.Error("登录失败，用户名: %s, 错误: %v", req.Username, err)
		sendResponse(w, http.StatusUnauthorized, false, "用户名或密码错误")
		return
	}
	if !user.Enable {
		utils.AuthLogger.Info("登录失败，账户已被禁用，用户名: %s", req.Username)
		sendResponse(w, http.StatusForbidden, false, "账户已被禁用")
		return

	}
	err = database.UpdateLoginTime(user.ID)
	if err != nil {
		utils.AuthLogger.Error("更新登录时间失败，用户ID: %d, 错误: %v", user.ID, err)
	}
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		utils.AuthLogger.Error("生成令牌失败，用户ID: %d, 错误: %v", user.ID, err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "生成令牌失败")
		return
	}
	utils.SuccessResponse(w, "登录成功", map[string]interface{}{
		"token":    token,
		"id":       user.ID,
		"role":     user.Role,
		"username": user.Username,
	})
	utils.AuthLogger.Info("登录成功，用户名: %s, 用户ID: %d", user.Username, user.ID)
}

// checkLogin 验证用户登录信息
// 参数 username: 用户名
// 参数 password: 密码
// 返回值 *models.User: 用户对象，如果验证成功
// 返回值 error: 错误信息，如果验证失败
func checkLogin(username, password string) (*models.User, error) {
	user, err := database.FindUserByUsername(username, password)
	// 1. 必须先处理错误或 nil，并立即返回
	if err != nil {
		utils.AuthLogger.Error("数据库查询出错: %v", err)
		return nil, err
	}
	return user, nil
}

// sendResponse 发送HTTP响应
// 设置响应头为JSON格式，写入指定状态码并编码响应数据
func sendResponse(w http.ResponseWriter, code int, success bool, message string) {
	if !success {
		utils.ErrorResponse(w, code, message)
		return
	}

	utils.SuccessResponse(w, message, nil)
}
