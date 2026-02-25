package service

import (
	"GoWork_7/internal/models"
	"GoWork_7/internal/repository"
	"GoWork_7/internal/utils"
	"errors"
)

// LoginService 登录业务服务
type LoginService struct {
	userRepo *repository.UserRepository
}

// NewLoginService 创建登录服务实例
func NewLoginService(userRepo *repository.UserRepository) *LoginService {
	return &LoginService{userRepo: userRepo}
}

// Login 处理登录业务逻辑
// 参数: username 用户名, password 密码
// 返回: *models.User 用户对象, string JWT令牌, error 错误信息
func (s *LoginService) Login(username, password string) (*models.User, string, error) {
	// 1. 获取用户信息
	user, err := s.userRepo.GetByUsernameAndPassword(username, password)
	if err != nil {
		return nil, "", err
	}

	// 2. 检查账号是否启用
	if !user.Enable {
		return nil, "", errors.New("ACCOUNT_DISABLED")
	}

	// 3. 更新登录时间
	_ = s.userRepo.UpdateLoginTime(user.ID)

	// 4. 生成 JWT Token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// GetUserByID 根据ID获取用户信息 (保留在登录服务中供相关逻辑使用)
func (s *LoginService) GetUserByID(id int64) (*models.User, error) {
	return s.userRepo.GetByID(id)
}
