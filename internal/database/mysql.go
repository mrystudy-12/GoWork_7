package database

import (
	"GoWork_7/internal/utils"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var (
	DB  *sql.DB
	err error
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
