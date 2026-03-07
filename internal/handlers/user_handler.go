package handlers

import (
	"GoWork_7/internal/models"
	"GoWork_7/internal/service"
	"GoWork_7/internal/utils"
	"encoding/json"
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
// @Summary 获取用户列表
// @Description 根据分页、关键词和状态获取用户列表
// @Tags 用户管理
// @Accept  json
// @Produce  json
// @Param   page     query    int     false  "页码 (默认 1)"
// @Param   limit    query    int     false  "每页数量 (默认 10)"
// @Param   keyword  query    string  false  "搜索关键词"
// @Param   status   query    string  false  "用户状态"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}} "获取成功"
// @Failure 405 {object} models.APIResponse "方法不允许"
// @Failure 500 {object} models.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /users [get]
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

	// 统一处理头像 URL 拼接
	for i := range users {
		users[i].Avatar = utils.FormatAvatarURL(r, users[i].Avatar)
	}

	utils.SuccessResponse(w, "查询成功", map[string]interface{}{
		"users": users,
		"total": total,
	})
}

// NewUser 创建新用户（仅管理员）
// @Summary 创建新用户
// @Description 管理员创建一个新的用户
// @Tags 用户管理
// @Accept  json
// @Produce  json
// @Param   request  body      object  true  "创建用户请求体"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}} "创建成功"
// @Failure 400 {object} models.APIResponse "无效的请求参数"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 500 {object} models.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /users [post]
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

// PutUser 修改用户信息 (RESTful: PUT /api/users/{id})
// @Summary 修改用户信息
// @Description 管理员或用户本人修改用户信息
// @Tags 用户管理
// @Accept  json
// @Produce  json
// @Param   id       path     int     true  "用户ID"
// @Param   request  body      models.User  true  "更新用户请求体"
// @Success 200 {object} models.APIResponse{data=models.User} "修改成功"
// @Failure 400 {object} models.APIResponse "无效的用户ID或请求参数"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 404 {object} models.APIResponse "找不到用户"
// @Failure 500 {object} models.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *UserHandler) PutUser(w http.ResponseWriter, r *http.Request) {
	operatorRole, _ := r.Context().Value("role").(string)
	operatorID, _ := r.Context().Value("userID").(int64)

	// 从 URL 路径中获取 ID
	idStr := r.PathValue("id")
	targetID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "无效的用户ID")
		return
	}

	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "无效的请求参数")
		return
	}
	u.ID = targetID // 强制使用 URL 中的 ID

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

	// 统一处理返回的头像 URL
	u.Avatar = utils.FormatAvatarURL(r, u.Avatar)

	utils.SuccessResponse(w, "修改成功", u)
}

// DeleteUser 删除用户 (RESTful: DELETE /api/users/{id})
// @Summary 删除用户
// @Description 管理员删除指定用户
// @Tags 用户管理
// @Accept  json
// @Produce  json
// @Param   id       path     int     true  "用户ID"
// @Success 200 {object} models.APIResponse{data=map[string]interface{}} "删除成功"
// @Failure 400 {object} models.APIResponse "无效的用户ID"
// @Failure 403 {object} models.APIResponse "权限不足"
// @Failure 500 {object} models.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value("role").(string)
	if role != "admin" {
		utils.ErrorResponse(w, http.StatusForbidden, "权限不足")
		return
	}

	// 从 URL 路径中获取 ID
	idStr := r.PathValue("id")
	finalID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "无效的用户ID")
		return
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
