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

// Server 结构体管理所有的 Handler 和 Middleware 依赖
type Server struct {
	authProvider    *middleware.AuthMiddlewareProvider
	loginHandler    *handlers.LoginHandler
	registerHandler *handlers.RegisterHandler
	userHandler     *handlers.UserHandler
	uploadHandler   *handlers.UploadHandler
}

// NewServer 初始化所有的依赖项并返回 Server 实例
func NewServer() *Server {
	// 1. 初始化仓储层和业务层依赖
	userRepo := repository.NewUserRepository(database.DB)
	userService := service.NewUserService(userRepo)
	loginService := service.NewLoginService(userRepo)
	registerService := service.NewRegisterService(userRepo)

	// 2. 初始化所有的 Handler 和 Middleware Provider
	return &Server{
		authProvider:    middleware.NewAuthMiddlewareProvider(userRepo),
		loginHandler:    handlers.NewLoginHandler(loginService),
		registerHandler: handlers.NewRegisterHandler(registerService),
		userHandler:     handlers.NewUserHandler(userService),
		uploadHandler:   handlers.NewUploadHandler(userService),
	}
}

// SetupRouter 启动路由分发器并返回包装了全局中间件的 Handler
func SetupRouter() http.Handler {
	s := NewServer()
	mux := http.NewServeMux()
	auth := s.authProvider.AuthMiddleware // 快捷引用中间件

	// 1. 注册公开路由 (无需鉴权)
	// 静态资源
	mux.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("view/html"))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("view/js"))))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("view/images"))))

	// 公开页面与接口
	mux.HandleFunc("/", welcome3)
	mux.HandleFunc("/login.html", welcome3)
	mux.HandleFunc("POST /api/auth/login", s.loginHandler.Login)
	mux.HandleFunc("POST /api/auth/register", s.registerHandler.Register)

	// 2. 注册受保护路由 (统一绑定 AuthMiddleware)
	// 受保护页面
	mux.Handle("GET /index.html", auth(http.HandlerFunc(welcome3)))
	mux.Handle("GET /userList.html", auth(http.HandlerFunc(welcome3)))

	// 用户管理接口
	mux.Handle("GET /api/users", auth(http.HandlerFunc(s.userHandler.GetAllUsers)))
	mux.Handle("POST /api/users", auth(http.HandlerFunc(s.userHandler.NewUser)))
	mux.Handle("PUT /api/users/{id}", auth(http.HandlerFunc(s.userHandler.PutUser)))
	mux.Handle("DELETE /api/users/{id}", auth(http.HandlerFunc(s.userHandler.DeleteUser)))
	mux.Handle("POST /api/uploads/avatar", auth(http.HandlerFunc(s.uploadHandler.UploadAvatar)))
	mux.Handle("POST /api/users/{id}/avatar", auth(http.HandlerFunc(s.uploadHandler.UploadAvatar)))

	// 3. 应用全局系统中间件链
	return middleware.Recover(middleware.CORS(middleware.Logging(mux)))
}
