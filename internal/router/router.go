package router

import (
	"GoWork_7/internal/handlers"
	"GoWork_7/internal/middleware"
	"GoWork_7/internal/utils"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

func welcome3(w http.ResponseWriter, r *http.Request) {

	// 保护 API 路径：如果是 /api/ 开头的请求走到这里，说明路径写错了
	if strings.HasPrefix(r.URL.Path, "/api/") {
		utils.ErrorResponse(w, http.StatusNotFound, "API 路径不存在，请检查大小写")
		return
	}

	// 精确匹配 HTML 页面
	switch r.URL.Path {
	case "/", "/login.html", "/welcome3":
		t, err := template.ParseFiles("view/html/login.html")
		fmt.Println(err)
		err = t.Execute(w, nil)
		fmt.Println(err)
	case "/index.html":
		t, _ := template.ParseFiles("view/html/index.html")
		t.Execute(w, nil)
	case "/userList.html":
		t, _ := template.ParseFiles("view/html/userList.html")
		t.Execute(w, nil)
	default:
		http.NotFound(w, r) // 其他未定义的路径直接 404
	}
}

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// 1. 静态资源 (建议通过 mux 挂载，确保路径统一)
	mux.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("view/html"))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("view/js"))))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("view/images"))))

	// 2. 基础页面路由
	mux.HandleFunc("/", welcome3)

	// 3. 公开接口 (无需 Token)
	mux.HandleFunc("/api/login", handlers.Login)       // 对应 login.js
	mux.HandleFunc("/api/register", handlers.Register) // 对应 register.js

	// 4. 受保护接口 (通过 AuthMiddleware 校验 Token)
	// 获取列表
	mux.Handle("/api/GetAllUsers", middleware.AuthMiddleware(http.HandlerFunc(handlers.GetAllUsers)))

	// 新增用户 (前端 userList.js 调用的是 /AddUser)
	mux.Handle("/api/AddUser", middleware.AuthMiddleware(http.HandlerFunc(handlers.NewUser)))

	// 修改用户 (前端 userList.js 调用的是 /PutUser)
	mux.Handle("/api/PutUser", middleware.AuthMiddleware(http.HandlerFunc(handlers.PutUser)))

	// 删除用户 (前端 userList.js 调用的是 /DeleteUser)
	mux.Handle("/api/DeleteUser", middleware.AuthMiddleware(http.HandlerFunc(handlers.DeleteUser)))

	// 上传头像
	mux.Handle("/api/upload-avatar", middleware.AuthMiddleware(http.HandlerFunc(handlers.UploadAvatar)))

	return mux
}
