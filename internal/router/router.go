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
	switch r.URL.Path {
	case "/", "/login.html", "/welcome3":
		t, err := template.ParseFiles("view/html/login.html")
		if err != nil {
			fmt.Println("Error parsing login.html:", err)
			return
		}
		t.Execute(w, nil)
	case "/index.html":
		t, err := template.ParseFiles("view/html/index.html")
		if err != nil {
			fmt.Println("Error parsing index.html:", err)
			return
		}
		t.Execute(w, nil)
	case "/userList.html":
		t, err := template.ParseFiles("view/html/userList.html")
		if err != nil {
			fmt.Println("Error parsing userList.html:", err)
			return
		}
		t.Execute(w, nil)
	default:
		http.NotFound(w, r)
	}
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

	authMiddleware := middleware.NewAuthMiddlewareProvider(userRepo)

	// 1. 静态资源
	mux.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("view/html"))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("view/js"))))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("view/images"))))

	// 2. 基础页面路由
	mux.HandleFunc("/", welcome3)

	// 3. 公开接口
	mux.HandleFunc("/api/login", loginHandler.Login)
	mux.HandleFunc("/api/register", registerHandler.Register)

	// 4. 受保护接口
	mux.Handle("/api/GetAllUsers", authMiddleware.AuthMiddleware(http.HandlerFunc(userHandler.GetAllUsers)))
	mux.Handle("/api/AddUser", authMiddleware.AuthMiddleware(http.HandlerFunc(userHandler.NewUser)))
	mux.Handle("/api/PutUser", authMiddleware.AuthMiddleware(http.HandlerFunc(userHandler.PutUser)))
	mux.Handle("/api/DeleteUser", authMiddleware.AuthMiddleware(http.HandlerFunc(userHandler.DeleteUser)))
	mux.Handle("/api/upload-avatar", authMiddleware.AuthMiddleware(http.HandlerFunc(handlers.UploadAvatar)))

	return mux
}
