package service

import (
	"GoWork_7/internal/repository"
)

// RegisterService 注册业务服务
type RegisterService struct {
	userRepo *repository.UserRepository
}

// NewRegisterService 创建注册服务实例
func NewRegisterService(userRepo *repository.UserRepository) *RegisterService {
	return &RegisterService{userRepo: userRepo}
}

// Register 处理注册业务逻辑
// 参数: username 用户名, password 密码
// 返回: int64 新用户ID, error 错误信息
func (s *RegisterService) Register(username, password string) (int64, error) {
	// 这里可以添加业务逻辑，比如校验用户名是否已存在、密码强度校验等
	return s.userRepo.Create(username, password)
}
