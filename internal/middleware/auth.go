package middleware

import (
	"GoWork_7/internal/database"
	"GoWork_7/internal/utils"
	"context"
	"net/http"
	"strings"
)

// setCorsHeaders 设置跨域头部
func setCorsHeaders(w http.ResponseWriter) {
	utils.SetCORSHeaders(w, "GET, POST, PUT, DELETE, OPTIONS")
}

// returnErrorWithCors 返回错误响应并设置跨域头部
func returnErrorWithCors(w http.ResponseWriter, code int, message string) {
	setCorsHeaders(w)
	http.Error(w, message, code)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			returnErrorWithCors(w, http.StatusUnauthorized, "Unauthorized: No token provided")
			utils.AuthLogger.Info("Token为空")
			return
		}
		// 2. 核心清洗逻辑：一定要把 "Bearer " 彻底剥离
		// 使用 strings.Fields 可以完美处理"Bearer<token>"或"Bearer <token>"
		parts := strings.Fields(authHeader)
		var tokenStr string

		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			// 这是标准情况："Bearer eyJ..."
			tokenStr = parts[1]
		} else if len(parts) == 1 && strings.HasPrefix(strings.ToLower(parts[0]), "bearer") {
			// 这是你前端之前那种粘连情况："BearereyJ..."
			tokenStr = authHeader[6:]
		} else {
			// 兜底：直接尝试解析整个字符串
			tokenStr = authHeader
		}
		// 3. 再次确保两端没有不可见字符
		tokenStr = strings.TrimSpace(tokenStr)

		claims, err := utils.ParseToken(tokenStr)
		if err != nil {
			returnErrorWithCors(w, http.StatusUnauthorized, "Unauthorized:Invalid token")
			utils.AuthLogger.Error("Token解析失败: %v", err)
			return
		}

		newRole, changed, active := checkUserPermissionFromDB(claims.ID, claims.Role)
		if !active {
			returnErrorWithCors(w, http.StatusForbidden, "账号已被禁用或不存在")
			utils.AuthLogger.Info("账号已被禁用或不存在，用户ID: %d", claims.ID)
			return
		}
		if changed {
			newToken, err := utils.GenerateToken(claims.ID, claims.Username, claims.Role)
			if err == nil {
				w.Header().Set("New-Token", newToken)
				w.Header().Set("Access-Control-Expose-Headers", "New-Token")
				claims.Role = newRole
				utils.AuthLogger.Info("生成新Token成功，用户ID: %d", claims.ID)
			} else {
				utils.AuthLogger.Error("生成新Token失败: %v", err)
			}
		}
		ctx := context.WithValue(r.Context(), "userID", claims.ID)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func checkUserPermissionFromDB(id int64, oldRole string) (string, bool, bool) {
	user, err := database.FindUserByID(id)
	if err != nil {
		// 查不到用户，视为不可用
		return oldRole, false, false
	}

	// 1. 首先检查禁用状态
	if !user.Enable {
		return oldRole, false, false // 账号已被封禁
	}

	// 2. 其次检查角色变更
	if user.Role != oldRole {
		return user.Role, true, true // 角色变了，但账号还是活跃的
	}

	// 3. 正常状态
	return oldRole, false, true
}
