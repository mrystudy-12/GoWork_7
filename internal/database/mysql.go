package database

import (
	"GoWork_7/internal/models"
	"GoWork_7/internal/utils"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var (
	ErrUserNotFound = errors.New("USER_NOT_FOUND")
	ErrDbInternal   = errors.New("DB_INTERNAL_ERROR")
	DB              *sql.DB
	err             error
)

func ConnectDB() {
	dsn := "root:231792@tcp(127.0.0.1:3306)/backstage"
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		utils.SystemLogger.Error("连接配置错误：%v", err)
		panic(err)
	}
	err = DB.Ping()
	if err != nil {
		utils.SystemLogger.Error("数据库无法访问：%v", err)
		panic(err)
	}
	utils.SystemLogger.Info("成功连接到MySQL数据库！")
}

func InsertUser(username, password string) (int64, error) {
	query := "INSERT INTO users(username,password) VALUES (?,?)"
	result, err := DB.Exec(query, username, password)
	if err != nil {
		utils.SystemLogger.Error("操作失败，请检查日志文件。错误: %v", err)
		return 0, err
	}
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	utils.SystemLogger.Info("插入成功，用户ID为：%d", lastInsertID)
	return lastInsertID, nil
}

func FindUserByUsername(username string, password string) (user *models.User, err error) {
	query := "SELECT id, username,password,role,status,avatar FROM users WHERE username = ? and password =?"

	u := &models.User{}
	var isAble string
	var avatar sql.NullString
	err = DB.QueryRow(query, username, password).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &isAble, &avatar)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound //没找到用户
		}
		return nil, err
	}
	if isAble == "enabled" {
		u.Enable = true
	} else if isAble == "disabled" {
		u.Enable = false
	}
	if avatar.Valid {
		u.Avatar = avatar.String
	}
	return u, nil
}

func DeleteUserByID(id int64) (rowsAffected int64, err error) {
	query := "DELETE FROM users WHERE id = ?"

	result, err := DB.Exec(query, id)
	if err != nil {
		// 如果 SQL 语法错误或连接断开，直接返回
		return 0, err
	}

	// 2. 从 result 中提取受影响的行数
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		// 注意：某些数据库驱动可能不支持 RowsAffected()
		return 0, err
	}

	return rowsAffected, nil
}

func FindUserByID(id int64) (*models.User, error) {
	u := &models.User{}
	query := "select id,username,password,role,status,avatar from Users where id =?"
	var isAble string
	var avatar sql.NullString
	err = DB.QueryRow(query, id).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &isAble, &avatar)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if isAble == "enabled" {
		u.Enable = true
	} else if isAble == "disabled" {
		u.Enable = false
	}
	if avatar.Valid {
		u.Avatar = avatar.String
	}
	return u, nil
}

// FetchUsersWithPagination 封装分页查询逻辑
// 输入：page 页码, limit 每页数量, keyword 搜索关键词, status 状态筛选
// 输出：用户切片, 总条数, 错误信息
func FetchUsersWithPagination(page, limit int, keyword, status string) ([]models.User, int, error) {
	// 1. 构建查询条件
	whereClause := ""
	var args []interface{}

	if keyword != "" {
		whereClause += " AND username LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	if status != "" {
		var statusValue string
		if status == "1" {
			statusValue = "enabled"
		} else {
			statusValue = "disabled"
		}
		whereClause += " AND status = ?"
		args = append(args, statusValue)
	}

	// 调整 WHERE 子句格式
	if whereClause != "" {
		whereClause = "WHERE " + whereClause[5:] // 移除开头的 " AND "
	}

	// 2. 查询总数据量
	var total int
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	var countErr error
	if len(args) > 0 {
		countErr = DB.QueryRow(countQuery, args...).Scan(&total)
	} else {
		countErr = DB.QueryRow(countQuery).Scan(&total)
	}
	if countErr != nil {
		return nil, 0, countErr
	}

	// 3. 执行分页查询
	offset := (page - 1) * limit
	query := `
		SELECT id, username, role, COALESCE(last_login, '1970-01-01 00:00:00'), status, avatar 
		FROM users 
		` + whereClause + `
		ORDER BY id ASC 
		LIMIT ? OFFSET ?`

	// 添加分页参数
	queryArgs := append(args, limit, offset)

	rows, err := DB.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// 4. 解析结果
	var users []models.User
	for rows.Next() {
		var u models.User
		var statusStr string
		var avatar sql.NullString
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.LastLogin, &statusStr, &avatar); err != nil {
			utils.SystemLogger.Error("扫描用户数据失败: %v", err)
			continue
		}

		// 转换状态
		if statusStr == "enabled" {
			u.Enable = true
		} else {
			u.Enable = false
		}

		// 处理可能为NULL的avatar字段
		if avatar.Valid {
			u.Avatar = avatar.String
		}

		users = append(users, u)
	}

	return users, total, nil
}

func UpdateUser(currentOperatorID int64, user *models.User) error {
	// 1. 权限防线：禁止用户修改自己的角色
	if currentOperatorID == user.ID {
		var currentRole string
		// 从数据库查询该用户当前的真实角色
		err := DB.QueryRow("SELECT role FROM Users WHERE id = ?", user.ID).Scan(&currentRole)
		if err != nil {
			return err
		}
		// 如果提交的角色与数据库现有的角色不一致，拒绝修改
		if user.Role != currentRole {
			return fmt.Errorf("提交失败：您不能修改自己的用户权限！")
		}
		fmt.Println(user.Enable)
		if !user.Enable {
			return fmt.Errorf("提交失败：你不能修改自己的用户状态")
		}
	}

	// 2. 状态转换逻辑
	var status string
	fmt.Println(user.Enable)
	if user.Enable {
		status = "enabled"
	} else {
		status = "disabled"
	}

	// 3. 动态 SQL 构建：实现“不输入新密码就不修改密码”
	var query string
	var args []interface{}

	if user.Password != "" {
		// 如果密码不为空，执行全量更新
		query = "UPDATE Users SET username=?, password=?, role=?, status=?, avatar=? WHERE id=?"
		args = []interface{}{user.Username, user.Password, user.Role, status, user.Avatar, user.ID}
	} else {
		// 如果密码为空，SQL 语句中剔除 password 字段
		query = "UPDATE Users SET username=?, role=?, status=?, avatar=? WHERE id=?"
		args = []interface{}{user.Username, user.Role, status, user.Avatar, user.ID}
	}

	// 4. 执行更新
	_, err := DB.Exec(query, args...)
	return err
}

func UpdateLoginTime(uid int64) error {
	query := "UPDATE Users SET last_Login = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := DB.Exec(query, uid)
	return err
}
