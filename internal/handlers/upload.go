package handlers

import (
	"GoWork_7/internal/utils"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// UploadAvatar 处理头像上传
func UploadAvatar(w http.ResponseWriter, r *http.Request) {
	// 检查请求方法
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 获取用户信息
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	// 获取用户名
	username, ok := r.Context().Value("username").(string)
	if !ok {
		utils.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 解析表单
	if err := r.ParseMultipartForm(2 * 1024 * 1024); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "File too large")
		return
	}

	// 获取文件
	file, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// 验证文件类型
	contentType := fileHeader.Header.Get("Content-Type")
	if !isAllowedType(contentType) {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid file type")
		return
	}

	// 创建上传目录（直接在view/images目录下）
	uploadDir := filepath.Join("view", "images")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create upload directory")
		return
	}

	// 生成文件名（使用用户ID_用户名.扩展名格式）
	ext := filepath.Ext(fileHeader.Filename)
	// 处理用户名，替换空格为下划线
	processedUsername := strings.ReplaceAll(username, " ", "_")
	fileName := fmt.Sprintf("%d_%s%s", userID, processedUsername, ext)
	filePath := filepath.Join(uploadDir, fileName)

	// 保存文件（直接覆盖旧文件）
	dst, err := os.Create(filePath)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := dst.ReadFrom(file); err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// 返回相对路径
	utils.SuccessResponse(w, "Upload success", map[string]interface{}{
		"path": fileName,
	})
}

// isAllowedType 检查文件类型是否允许
func isAllowedType(contentType string) bool {
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}
