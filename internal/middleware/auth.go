package middleware

import (
	"GoWork_7/internal/repository"
	"GoWork_7/internal/utils"
	"context"
	"net/http"
	"strings"
)

// AuthMiddlewareProvider 认证中间件提供者
type AuthMiddlewareProvider struct {
	userRepo *repository.UserRepository
}

// NewAuthMiddlewareProvider 创建认证中间件提供者实例
func NewAuthMiddlewareProvider(userRepo *repository.UserRepository) *AuthMiddlewareProvider {
	return &AuthMiddlewareProvider{userRepo: userRepo}
}

// AuthMiddleware 核心认证中间件
// 负责校验请求头中的 JWT Token，验证用户身份和权限
func (p *AuthMiddlewareProvider) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 获取 Authorization 请求头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.SetCORSHeaders(w, "GET, POST, PUT, DELETE, OPTIONS")
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		// 2. 清洗 Token 字符串，支持 "Bearer <token>" 格式
		parts := strings.Fields(authHeader)
		var tokenStr string
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			tokenStr = parts[1]
		} else {
			tokenStr = authHeader
		}
		tokenStr = strings.TrimSpace(tokenStr)

		// 3. 解析并校验 Token
		claims, err := utils.ParseToken(tokenStr)
		if err != nil {
			utils.SetCORSHeaders(w, "GET, POST, PUT, DELETE, OPTIONS")
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// 4. 二次校验：检查数据库中用户状态和角色是否发生变更
		newRole, changed, active := p.checkUserPermissionFromDB(claims.ID, claims.Role)
		if !active {
			utils.SetCORSHeaders(w, "GET, POST, PUT, DELETE, OPTIONS")
			http.Error(w, "账号已被禁用或不存在", http.StatusForbidden)
			return
		}

		// 5. 如果角色发生变更，自动下发新 Token (实现无缝角色切换)
		if changed {
			newToken, err := utils.GenerateToken(claims.ID, claims.Username, newRole)
			if err == nil {
				w.Header().Set("New-Token", newToken)
				w.Header().Set("Access-Control-Expose-Headers", "New-Token")
				claims.Role = newRole
			}
		}

		// 6. 将用户信息注入 Context，供后续 Handler 使用
		ctx := context.WithValue(r.Context(), "userID", claims.ID)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "username", claims.Username)

		// 7. 继续执行下一个处理器
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// checkUserPermissionFromDB 校验用户在数据库中的实时状态
// 返回值: 实时角色, 角色是否变更, 账号是否活跃
func (p *AuthMiddlewareProvider) checkUserPermissionFromDB(id int64, oldRole string) (string, bool, bool) {
	user, err := p.userRepo.GetByID(id)
	if err != nil {
		return oldRole, false, false
	}

	// 检查是否被禁用
	if !user.Enable {
		return oldRole, false, false
	}

	// 检查角色是否变更
	if user.Role != oldRole {
		return user.Role, true, true
	}

	return oldRole, false, true
}
