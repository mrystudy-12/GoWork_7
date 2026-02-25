package service

import (
	"GoWork_7/internal/models"
	"GoWork_7/internal/repository"
)

// UserService 用户管理业务服务
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetAllUsers 获取所有用户（分页+搜索）
func (s *UserService) GetAllUsers(page, limit int, keyword, status string) ([]models.User, int, error) {
	return s.userRepo.FetchWithPagination(page, limit, keyword, status)
}

// CreateUser 创建新用户
func (s *UserService) CreateUser(username, password string) (int64, error) {
	return s.userRepo.Create(username, password)
}

// GetUserByID 根据ID获取用户信息
func (s *UserService) GetUserByID(id int64) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *models.User) error {
	return s.userRepo.Update(user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id int64) (int64, error) {
	return s.userRepo.Delete(id)
}
