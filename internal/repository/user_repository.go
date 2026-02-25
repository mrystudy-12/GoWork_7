package repository

import (
	"GoWork_7/internal/models"
	"database/sql"
	"errors"
)

var (
	// ErrUserNotFound 用户不存在错误
	ErrUserNotFound = errors.New("USER_NOT_FOUND")
)

// UserRepository 用户数据访问仓库
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建新用户
// 参数: username 用户名, password 密码
// 返回: int64 新用户ID, error 错误信息
func (r *UserRepository) Create(username, password string) (int64, error) {
	query := "INSERT INTO users(username,password) VALUES (?,?)"
	result, err := r.db.Exec(query, username, password)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetByUsernameAndPassword 根据用户名和密码获取用户
// 参数: username 用户名, password 密码
// 返回: *models.User 用户对象, error 错误信息
func (r *UserRepository) GetByUsernameAndPassword(username, password string) (*models.User, error) {
	query := "SELECT id, username, password, role, status, avatar FROM users WHERE username = ? AND password = ?"
	u := &models.User{}
	var statusStr string
	var avatar sql.NullString

	err := r.db.QueryRow(query, username, password).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &statusStr, &avatar)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	r.mapUserStatus(u, statusStr, avatar)
	return u, nil
}

// GetByID 根据用户ID获取用户
// 参数: id 用户ID
// 返回: *models.User 用户对象, error 错误信息
func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	query := "SELECT id, username, password, role, status, avatar FROM users WHERE id = ?"
	u := &models.User{}
	var statusStr string
	var avatar sql.NullString

	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &statusStr, &avatar)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	r.mapUserStatus(u, statusStr, avatar)
	return u, nil
}

// Delete 根据用户ID删除用户
// 参数: id 用户ID
// 返回: int64 影响行数, error 错误信息
func (r *UserRepository) Delete(id int64) (int64, error) {
	query := "DELETE FROM users WHERE id = ?"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// FetchWithPagination 分页获取用户列表
// 参数: page 页码, limit 每页条数, keyword 关键词, status 状态筛选
// 返回: []models.User 用户切片, int 总记录数, error 错误信息
func (r *UserRepository) FetchWithPagination(page, limit int, keyword, status string) ([]models.User, int, error) {
	whereClause := ""
	var args []interface{}

	if keyword != "" {
		whereClause += " AND username LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	if status != "" {
		statusValue := "disabled"
		if status == "1" {
			statusValue = "enabled"
		}
		whereClause += " AND status = ?"
		args = append(args, statusValue)
	}

	if whereClause != "" {
		whereClause = "WHERE " + whereClause[5:]
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := `
		SELECT id, username, role, COALESCE(last_login, '1970-01-01 00:00:00'), status, avatar 
		FROM users 
		` + whereClause + `
		ORDER BY id ASC 
		LIMIT ? OFFSET ?`

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var statusStr string
		var avatar sql.NullString
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.LastLogin, &statusStr, &avatar); err != nil {
			continue
		}
		r.mapUserStatus(&u, statusStr, avatar)
		users = append(users, u)
	}
	return users, total, nil
}

// Update 更新用户信息
// 参数: user 用户对象
// 返回: error 错误信息
func (r *UserRepository) Update(user *models.User) error {
	var status string
	if user.Enable {
		status = "enabled"
	} else {
		status = "disabled"
	}

	var query string
	var args []interface{}

	if user.Password != "" {
		query = "UPDATE users SET username=?, password=?, role=?, status=?, avatar=? WHERE id=?"
		args = []interface{}{user.Username, user.Password, user.Role, status, user.Avatar, user.ID}
	} else {
		query = "UPDATE users SET username=?, role=?, status=?, avatar=? WHERE id=?"
		args = []interface{}{user.Username, user.Role, status, user.Avatar, user.ID}
	}

	_, err := r.db.Exec(query, args...)
	return err
}

// UpdateLoginTime 更新最后登录时间
// 参数: uid 用户ID
// 返回: error 错误信息
func (r *UserRepository) UpdateLoginTime(uid int64) error {
	query := "UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := r.db.Exec(query, uid)
	return err
}

// mapUserStatus 映射用户状态及头像
func (r *UserRepository) mapUserStatus(u *models.User, statusStr string, avatar sql.NullString) {
	u.Enable = (statusStr == "enabled")
	if avatar.Valid {
		u.Avatar = avatar.String
	}
}
