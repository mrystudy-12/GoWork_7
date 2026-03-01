package router

import (
	"GoWork_7/internal/database"
	"GoWork_7/internal/handlers"
	"GoWork_7/internal/middleware"
	"GoWork_7/internal/repository"
	"GoWork_7/internal/service"
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
	t, err := template.ParseFiles("view/html/login.html")
	if err != nil {
		fmt.Println("Error parsing login.html:", err)
		return
	}
	t.Execute(w, nil)
}

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// 初始化依赖
	userRepo := repository.NewUserRepository(database.DB)

	loginService := service.NewLoginService(userRepo)
	loginHandler := handlers.NewLoginHandler(loginService)

	registerService := service.NewRegisterService(userRepo)
	registerHandler := handlers.NewRegisterHandler(registerService)

	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	uploadHandler := handlers.NewUploadHandler(userService)

	authMiddleware := middleware.NewAuthMiddlewareProvider(userRepo)

	// 1. 静态资源
	mux.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("view/html"))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("view/js"))))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("view/images"))))

	// 2. 基础页面路由
	mux.HandleFunc("/", welcome3)

	// 3. 认证相关接口 (Restful: /api/auth/...)
	mux.HandleFunc("POST /api/auth/login", loginHandler.Login)
	mux.HandleFunc("POST /api/auth/register", registerHandler.Register)

	// 4. 用户资源接口 (Restful: /api/users)
	// 获取用户列表
	mux.Handle("GET /api/users", authMiddleware.AuthMiddleware(http.HandlerFunc(userHandler.GetAllUsers)))
	// 新增用户
	mux.Handle("POST /api/users", authMiddleware.AuthMiddleware(http.HandlerFunc(userHandler.NewUser)))
	// 修改用户 (使用路径参数 {id})
	mux.Handle("PUT /api/users/{id}", authMiddleware.AuthMiddleware(http.HandlerFunc(userHandler.PutUser)))
	// 删除用户 (使用路径参数 {id})
	mux.Handle("DELETE /api/users/{id}", authMiddleware.AuthMiddleware(http.HandlerFunc(userHandler.DeleteUser)))
	// 上传头像 (通用接口，支持新建用户时的临时上传)
	mux.Handle("POST /api/uploads/avatar", authMiddleware.AuthMiddleware(http.HandlerFunc(uploadHandler.UploadAvatar)))
	// 上传头像 (特定用户接口)
	mux.Handle("POST /api/users/{id}/avatar", authMiddleware.AuthMiddleware(http.HandlerFunc(uploadHandler.UploadAvatar)))

	return mux
}
