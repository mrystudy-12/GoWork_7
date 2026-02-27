package handlers

import (
	"GoWork_7/internal/service"
	"GoWork_7/internal/utils"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// UploadHandler 专门处理文件上传的控制器
type UploadHandler struct {
	userService *service.UserService
}

// NewUploadHandler 创建上传控制器实例
func NewUploadHandler(userService *service.UserService) *UploadHandler {
	return &UploadHandler{userService: userService}
}

// UploadAvatar 处理头像上传 (RESTful: POST /api/users/{id}/avatar 或 POST /api/uploads/avatar)
// 实现逻辑：通过 ID 查询数据库获取用户名，生成 ID_用户名.扩展名 格式的文件名
func (h *UploadHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	// 1. 检查请求方法
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 2. 获取 URL 中的用户 ID (如果是通用上传接口，则为空)
	idStr := r.PathValue("id")
	var targetID int64
	var isGeneric bool

	if idStr == "" {
		isGeneric = true
	} else {
		var err error
		targetID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			utils.ErrorResponse(w, http.StatusBadRequest, "无效的用户ID")
			return
		}
	}

	// 3. 获取当前登录用户信息（从中间件注入的 Context）
	operatorID, ok := r.Context().Value("userID").(int64)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	operatorRole, _ := r.Context().Value("role").(string)

	// 4. 权限校验
	if !isGeneric {
		// 特定用户上传：只能上传自己的头像，除非是管理员
		if operatorID != targetID && operatorRole != "admin" {
			utils.ErrorResponse(w, http.StatusForbidden, "无权修改他人头像")
			return
		}
	}

	// 5. 解析表单 (2MB 限制)
	if err := r.ParseMultipartForm(2 * 1024 * 1024); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "File too large")
		return
	}

	// 6. 获取文件
	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// 7. 验证文件类型
	contentType := fileHeader.Header.Get("Content-Type")
	if !h.isAllowedType(contentType) {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid file type")
		return
	}

	// 8. 创建上传目录
	uploadDir := filepath.Join("view", "images")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create upload directory")
		return
	}

	// 9. 生成文件名 (核心逻辑：ID + 用户名)
	ext := filepath.Ext(fileHeader.Filename)
	var fileName string
	if isGeneric {
		// 通用上传：使用时间戳+随机数
		fileName = fmt.Sprintf("temp_%d%s", time.Now().UnixNano(), ext)
	} else {
		// 特定用户上传：通过 ID 查询数据库获取用户名
		targetUser, err := h.userService.GetUserByID(targetID)
		username := "user"
		if err == nil && targetUser.Username != "" {
			// 处理用户名，替换空格为下划线
			username = strings.ReplaceAll(targetUser.Username, " ", "_")
		}
		// 格式：ID_用户名.扩展名
		fileName = fmt.Sprintf("%d_%s%s", targetID, username, ext)
	}
	filePath := filepath.Join(uploadDir, fileName)

	// 10. 保存文件 (覆盖旧文件)
	dst, err := os.Create(filePath)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// 11. 返回文件名
	utils.SuccessResponse(w, "Upload success", map[string]interface{}{
		"path": fileName,
	})
}

// isAllowedType 检查文件类型是否允许
func (h *UploadHandler) isAllowedType(contentType string) bool {
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}
