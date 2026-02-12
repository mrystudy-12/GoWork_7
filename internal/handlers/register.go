package handlers

import (
	"GoWork_7/internal/database"
	"GoWork_7/internal/utils"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {
	// 1. 仅允许 POST
	if r.Method != http.MethodPost {
		utils.AuthLogger.Info("注册请求方法错误: %s", r.Method)
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// 2. 解析表单
	if err := r.ParseForm(); err != nil {
		utils.AuthLogger.Error("解析注册表单失败: %v", err)
		utils.ErrorResponse(w, http.StatusBadRequest, "解析表单失败")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// 3. 服务端验证：这是你的“最后一道防线”
	// 即使黑客用工具绕过前端 JS 发送了 100 位的密码，这里也会把它挡掉
	if username == "" || len(password) != 6 {
		utils.AuthLogger.Info("注册验证失败，用户名: %s, 密码长度: %d", username, len(password))
		utils.ErrorResponse(w, http.StatusBadRequest, "格式错误：用户名不能为空且密码必须为6位")
		return
	}

	// 4. 调用 SQL 插入函数
	// 注意：你之前的 InsertUser 返回 (int64, error)，这里必须同时接收两个值
	uid, err := database.InsertUser(username, password)
	if err != nil {
		// 这里的 err 通常是数据库报错，比如用户名重复
		utils.AuthLogger.Error("注册失败，用户名: %s, 错误: %v", username, err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "注册失败：用户名可能已被占用")
		return
	}

	// 5. 生成 Token
	// 此时我们已经确定插入成功了，可以安全地发放令牌
	token, err := utils.GenerateToken(uid, username, "common")
	if err != nil {
		utils.AuthLogger.Error("生成令牌失败，用户ID: %d, 错误: %v", uid, err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "令牌生成失败")
		return
	}

	// 6. 返回结果：带上刚刚生成的 uid 告诉前端“插入成功了”
	utils.AuthLogger.Info("注册成功，用户名: %s, 用户ID: %d", username, uid)
	utils.SuccessResponse(w, "注册成功", map[string]interface{}{
		"user_id": uid,
		"token":   token,
	})

}
