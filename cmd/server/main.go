package main

import (
	"GoWork_7/internal/database"
	"GoWork_7/internal/router"
	"GoWork_7/internal/utils"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 初始化日志记录器
	if err := utils.InitLoggers(); err != nil {
		log.Fatalf("初始化日志记录器失败: %v", err)
	}
	utils.SystemLogger.Info("日志记录器初始化成功")

	// 连接数据库
	utils.SystemLogger.Info("正在连接数据库...")
	database.ConnectDB()
	defer database.DB.Close()
	utils.SystemLogger.Info("数据库连接成功")

	// 设置路由
	utils.SystemLogger.Info("正在设置路由...")
	r := router.SetupRouter()
	utils.SystemLogger.Info("路由设置成功")

	// 启动服务器
	addr := ":8090"
	utils.SystemLogger.Info("服务器已启动，监听地址：http://localhost:%s", addr)
	fmt.Printf("服务器已启动，监听地址：http://localhost:%s\n", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		utils.SystemLogger.Error("服务器启动失败: %v", err)
		fmt.Printf("服务器启动失败:%v\n", err)
		log.Fatal(err)
	}
}
