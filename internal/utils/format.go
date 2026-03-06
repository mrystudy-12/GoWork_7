package utils

import (
	"fmt"
	"net/http"
	"strings"
)

// FormatAvatarURL 统一处理头像路径拼接逻辑
func FormatAvatarURL(r *http.Request, avatarName string) string {
	if avatarName == "" {
		return ""
	}

	// 如果已经是完整路径（以 http 开头）则不处理，直接返回
	if strings.HasPrefix(avatarName, "http") {
		return avatarName
	}

	// 1. 获取协议 (HTTP/HTTPS)
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}

	// 2. 获取当前主机名 (例如 localhost:8090)
	host := r.Host
	if host == "" {
		host = "localhost:8090" // 默认兜底配置
	}

	// 3. 拼装完整路径：协议://域名/静态资源前缀/文件名
	return fmt.Sprintf("%s://%s/images/%s", protocol, host, avatarName)
}
