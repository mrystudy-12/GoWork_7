package handlers

import (
	"GoWork_7/internal/database"
	"GoWork_7/internal/models"
	"GoWork_7/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// 1. 限制请求方法为 GET
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// 2. 解析参数
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	keyword := r.URL.Query().Get("keyword")
	status := r.URL.Query().Get("status")
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	// 3. 调用封装好的 SQL 函数
	users, total, err := database.FetchUsersWithPagination(page, limit, keyword, status)
	if err != nil {
		utils.UserLogger.Error("数据库查询失败: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "数据库查询失败")
		return
	}

	// 4. 处理用户头像URL
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:8090"
	}

	// 遍历用户列表，拼接完整的头像URL
	for i := range users {
		if users[i].Avatar != "" {
			users[i].Avatar = fmt.Sprintf("%s://%s/images/%s", protocol, host, users[i].Avatar)
		}
	}

	utils.UserLogger.Debug("查询用户列表成功，总数: %d", total)
	// 5. 发送响应
	utils.SuccessResponse(w, "查询成功", map[string]interface{}{
		"users": users,
		"total": total,
	})
}

func NewUser(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value("role").(string)
	if role != "admin" {
		utils.ErrorResponse(w, http.StatusForbidden, "权限不足，只有管理员可以创建用户")
		return
	}
	// 1. 限制请求方法
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// 2. 解析 JSON
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 3. 调用 mysql.go 中的 InsertUser (复用 SQL)
	lastID, err := database.InsertUser(data.Username, data.Password)
	if err != nil {
		// 可以在这里根据错误类型返回不同的提示，比如用户名已存在
		utils.UserLogger.Error("插入数据库失败: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "插入数据库失败")
		return
	}
	utils.UserLogger.Info("新建用户成功，用户ID: %d, 用户名: %s", lastID, data.Username)

	// 4. 返回 JSON 给前端
	utils.SuccessResponse(w, "新建成功", map[string]interface{}{
		"id": lastID, // 传回 ID
	})
}

func PutUser(w http.ResponseWriter, r *http.Request) {
	operatorRole, _ := r.Context().Value("role").(string)
	operatorID, _ := r.Context().Value("userID").(int64)

	// 1. 限制请求方法为 PUT
	if r.Method != http.MethodPut {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	// 2. 解析前端传来的 JSON 到结构体
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// A. 查出数据库中该用户目前的真实信息
	targetUser, err := database.FindUserByID(u.ID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "找不到要修改的目标用户")
		return
	}

	// B. 核心判定逻辑
	if operatorRole == "admin" {
		// 如果当前操作者是管理员，且目标也是管理员，但 ID 不同（说明在改别人）
		if targetUser.Role == "admin" && operatorID != u.ID {
			utils.ErrorResponse(w, http.StatusForbidden, "管理员禁止修改其他管理员的信息")
			return
		}
	} else {
		// 如果是普通用户，必须 ID 一致才能改（只能改自己）
		if operatorID != u.ID {
			utils.ErrorResponse(w, http.StatusForbidden, "普通用户无权修改他人信息")
			return
		}
	}
	// 3. 调用底层的 UpdateUser 函数执行数据库修改
	// 注意：这里传入 u 的指针 &u
	err = database.UpdateUser(operatorID, &u)

	// 4. 根据数据库操作结果返回不同的响应包
	if err != nil {
		utils.UserLogger.Error("数据库修改失败: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "数据库修改失败: "+err.Error())
		return
	}
	utils.UserLogger.Info("修改用户成功，用户ID: %d, 用户名: %s", u.ID, u.Username)

	// 5. 修改成功
	utils.SuccessResponse(w, "修改成功", u)

}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// 1. 设置跨域头 (解决“无法连接服务器”的核心)
	role, _ := r.Context().Value("role").(string)
	if role != "admin" {
		utils.ErrorResponse(w, http.StatusForbidden, "权限不足，只有管理员可以删除用户")
		return
	}
	utils.SetCORSHeaders(w, "DELETE, OPTIONS")

	// 2. 处理浏览器预检请求 (OPTIONS)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. 严格限制只能使用 DELETE 请求
	if r.Method != http.MethodDelete {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// 4. 解析 Body (使用 interface{} 兼容前端传来的数字或字符串 ID)
	var input struct {
		ID interface{} `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 5. 转换 ID 类型 (将字符串 '12' 转换为 int64)
	var finalID int64
	switch v := input.ID.(type) {
	case string:
		finalID, _ = strconv.ParseInt(v, 10, 64)
	case float64:
		finalID = int64(v)
	default:
		// 如果类型完全不对
		utils.ErrorResponse(w, http.StatusBadRequest, "ID 格式错误")
		return
	}

	// 6. 管理员不能删除自己
	userID, _ := r.Context().Value("userID").(int64)
	if finalID == userID {
		utils.ErrorResponse(w, http.StatusForbidden, "管理员不能删除自己的账号")
		return
	}

	// 7. 调用数据库操作
	rowsAffected, err := database.DeleteUserByID(finalID)

	// 7. 返回 JSON 响应
	if err != nil {
		utils.UserLogger.Error("数据库操作失败: %v", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, "数据库操作失败")
		return
	}
	utils.UserLogger.Info("删除用户成功，用户ID: %d, 影响行数: %d", finalID, rowsAffected)

	utils.SuccessResponse(w, "删除成功", map[string]interface{}{
		"affected_rows": rowsAffected,
	})

}
