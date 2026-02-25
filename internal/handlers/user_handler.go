package handlers

import (
	"GoWork_7/internal/models"
	"GoWork_7/internal/service"
	"GoWork_7/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// UserHandler 用户模块控制器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户控制器实例
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetAllUsers 获取所有用户列表（分页+搜索）
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// 解析分页和筛选参数
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

	users, total, err := h.userService.GetAllUsers(page, limit, keyword, status)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "数据库查询失败")
		return
	}

	// 处理头像 URL 拼接
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:8090"
	}

	for i := range users {
		if users[i].Avatar != "" {
			users[i].Avatar = fmt.Sprintf("%s://%s/images/%s", protocol, host, users[i].Avatar)
		}
	}

	utils.SuccessResponse(w, "查询成功", map[string]interface{}{
		"users": users,
		"total": total,
	})
}

// NewUser 创建新用户（仅管理员）
func (h *UserHandler) NewUser(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value("role").(string)
	if role != "admin" {
		utils.ErrorResponse(w, http.StatusForbidden, "权限不足")
		return
	}

	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	lastID, err := h.userService.CreateUser(data.Username, data.Password)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "插入数据库失败")
		return
	}

	utils.SuccessResponse(w, "新建成功", map[string]interface{}{"id": lastID})
}

// PutUser 修改用户信息
func (h *UserHandler) PutUser(w http.ResponseWriter, r *http.Request) {
	operatorRole, _ := r.Context().Value("role").(string)
	operatorID, _ := r.Context().Value("userID").(int64)

	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	targetUser, err := h.userService.GetUserByID(u.ID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusNotFound, "找不到用户")
		return
	}

	// 权限检查逻辑
	if operatorRole == "admin" {
		if targetUser.Role == "admin" && operatorID != u.ID {
			utils.ErrorResponse(w, http.StatusForbidden, "禁止修改其他管理员")
			return
		}
	} else {
		if operatorID != u.ID {
			utils.ErrorResponse(w, http.StatusForbidden, "无权修改他人信息")
			return
		}
	}

	if err := h.userService.UpdateUser(&u); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "修改失败")
		return
	}

	utils.SuccessResponse(w, "修改成功", u)
}

// DeleteUser 删除用户（仅管理员）
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value("role").(string)
	if role != "admin" {
		utils.ErrorResponse(w, http.StatusForbidden, "权限不足")
		return
	}

	var input struct {
		ID interface{} `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	var finalID int64
	switch v := input.ID.(type) {
	case string:
		finalID, _ = strconv.ParseInt(v, 10, 64)
	case float64:
		finalID = int64(v)
	}

	operatorID, _ := r.Context().Value("userID").(int64)
	if finalID == operatorID {
		utils.ErrorResponse(w, http.StatusForbidden, "不能删除自己")
		return
	}

	affected, err := h.userService.DeleteUser(finalID)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "删除失败")
		return
	}

	utils.SuccessResponse(w, "删除成功", map[string]interface{}{"affected_rows": affected})
}
