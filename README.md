**项目概览**

- 基于 Go 标准库 net/http 的用户管理后台，提供登录、注册、用户增删改查、头像上传等接口
- 使用 JWT 进行认证与授权，区分 admin 与普通用户权限
- 前端静态资源通过内置服务器直接提供，含登录、注册、用户列表页面
- 采用分层结构：cmd、internal(database/handlers/middleware/models/router/utils)、view

**运行方式**

- 本地启动：服务默认监听 8090
  - 入口：[main.go](file:///D:/GoWork_7/cmd/server/main.go#L12-L39)
- 静态资源访问：
  - HTML: http://localhost:8090/html/...
  - JS: http://localhost:8090/js/...
  - 图片: http://localhost:8090/images/...
  - 页面路由：/、/login.html、/index.html、/userList.html 映射在 [router.go](file:///D:/GoWork_7/internal/router/router.go#L13-L37)
- Docker Compose：
  - 文件：[docker-compose.yml](file:///D:/GoWork_7/docker-compose.yml)
  - 声明了 app(8091) 与 MySQL(3306) 服务与初始化 SQL；与当前代码中的默认数据库连接不一致，使用前请统一

**目录结构**

- 后端
  - 入口：[main.go](file:///D:/GoWork_7/cmd/server/main.go)
  - 路由：[router.go](file:///D:/GoWork_7/internal/router/router.go#L39-L70)
  - 中间件(JWT/CORS)：[auth.go](file:///D:/GoWork_7/internal/middleware/auth.go)
  - 业务处理：登录/注册/用户/头像上传
    - [login.go](file:///D:/GoWork_7/internal/handlers/login.go)
    - [register.go](file:///D:/GoWork_7/internal/handlers/register.go)
    - [user.go](file:///D:/GoWork_7/internal/handlers/user.go)
    - [upload.go](file:///D:/GoWork_7/internal/handlers/upload.go)
  - 数据层：MySQL 连接与 SQL
    - [mysql.go](file:///D:/GoWork_7/internal/database/mysql.go)
    - 初始化 SQL：[init.sql](file:///D:/GoWork_7/init.sql)
  - 模型/响应格式：
    - [user.go](file:///D:/GoWork_7/internal/models/user.go)
  - 工具：JWT/日志/CORS/响应
    - [jwt.go](file:///D:/GoWork_7/internal/utils/jwt.go)
    - [logger.go](file:///D:/GoWork_7/internal/utils/logger.go)
    - [cors.go](file:///D:/GoWork_7/internal/utils/cors.go)
    - [error.go](file:///D:/GoWork_7/internal/utils/error.go)
- 前端
  - 模板与脚本：[view/html](file:///D:/GoWork_7/view/html)、[view/js](file:///D:/GoWork_7/view/js)、[view/images](file:///D:/GoWork_7/view/images)

**API 设计**

- 登录 POST /api/login
  - 认证：公开
  - 请求：JSON { username, password }
  - 响应：{ token, id, role, username }
  - 实现：[login.go](file:///D:/GoWork_7/internal/handlers/login.go#L11-L55)
- 注册 POST /api/register
  - 认证：公开
  - 请求：Form username、password(必须6位)
  - 响应：{ user_id, token }
  - 实现：[register.go](file:///D:/GoWork_7/internal/handlers/register.go#L9-L59)
- 获取用户列表 GET /api/GetAllUsers
  - 认证：Bearer Token
  - 请求参数：page, limit, keyword, status(1=enabled / 其他=disabled)
  - 响应：{ users, total }；头像字段返回完整 URL
  - 实现：[user.go:GetAllUsers](file:///D:/GoWork_7/internal/handlers/user.go#L13-L63)
- 新增用户 POST /api/AddUser
  - 认证：Bearer Token；admin 才可
  - 请求：JSON { username, password }
  - 响应：{ id }
  - 实现：[user.go:NewUser](file:///D:/GoWork_7/internal/handlers/user.go#L65-L101)
- 修改用户 PUT /api/PutUser
  - 认证：Bearer Token
  - 权限规则：
    - admin 不允许修改其他 admin
    - 普通用户只能修改自己
  - 请求：JSON User 对象(密码为空则不改密码)
  - 响应：修改后的用户对象
  - 实现：[user.go:PutUser](file:///D:/GoWork_7/internal/handlers/user.go#L103-L155)、[mysql.go:UpdateUser](file:///D:/GoWork_7/internal/database/mysql.go#L205-L250)
- 删除用户 DELETE /api/DeleteUser
  - 认证：Bearer Token；admin 才可，且不能删除自己
  - 请求：JSON { id }，id 可为字符串或数字
  - 响应：{ affected_rows }
  - 实现：[user.go:DeleteUser](file:///D:/GoWork_7/internal/handlers/user.go#L157-L222)
- 上传头像 POST /api/upload-avatar
  - 认证：Bearer Token
  - 请求：multipart/form-data，字段名 avatar
  - 响应：{ path }，文件保存于 view/images
  - 实现：[upload.go](file:///D:/GoWork_7/internal/handlers/upload.go#L12-L86)

**认证与授权**

- Token 签发与解析：
  - 过期时间 30 分钟，包含 ID/Username/Role
  - 代码：[jwt.go](file:///D:/GoWork_7/internal/utils/jwt.go#L17-L31)、[jwt.go](file:///D:/GoWork_7/internal/utils/jwt.go#L33-L44)
- 中间件行为：
  - 支持 "Bearer token" 与粘连 "Bearertoken" 格式
  - 请求上下文注入 userID、role、username
  - 当角色变更时返回 New-Token 响应头
  - 代码：[auth.go](file:///D:/GoWork_7/internal/middleware/auth.go#L22-L76)

**响应与跨域**

- 统一响应结构：
  - models.APIResponse { success, code, message, data }
  - 代码：[error.go](file:///D:/GoWork_7/internal/utils/error.go#L9-L20)、[error.go](file:///D:/GoWork_7/internal/utils/error.go#L22-L33)
- CORS：
  - 允许方法：GET, POST, PUT, DELETE, OPTIONS
  - 暴露头：New-Token
  - 代码：[cors.go](file:///D:/GoWork_7/internal/utils/cors.go#L5-L12)

**数据库**

- 连接配置(开发默认)：root:231792@tcp(127.0.0.1:3306)/backstage
  - 代码：[mysql.go:ConnectDB](file:///D:/GoWork_7/internal/database/mysql.go#L19-L32)
- 表结构与初始化：
  - users 表字段：id, username, password, last_login, role, status, avatar, created_at, updated_at
  - 默认插入 admin 用户，密码 123456
  - 文件：[init.sql](file:///D:/GoWork_7/init.sql)

**日志**

- 模块化日志器：auth/system/user 三类
- 日志输出同时写文件并打印到控制台
- 代码：[logger.go](file:///D:/GoWork_7/internal/utils/logger.go)

**静态资源**

- 资源映射：
  - /html/ → view/html
  - /js/ → view/js
  - /images/ → view/images
  - 代码：[router.go](file:///D:/GoWork_7/internal/router/router.go#L42-L46)

**示例请求**

- 登录获取 Token

```bash
curl -X POST http://localhost:8090/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'
```

- 获取用户列表

```bash
curl -X GET "http://localhost:8090/api/GetAllUsers?
page=1&limit=10" \
  -H "Authorization: Bearer <token>"
```

- 新增用户(admin)

```bash
curl -X POST http://localhost:8090/api/AddUser \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"654321"}'
```

- 修改用户

```bash
curl -X PUT http://localhost:8090/api/PutUser \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"id":2,"username":"alice","role":"user",
  "enable":true}'
```

- 删除用户(admin)

```bash
curl -X DELETE http://localhost:8090/api/DeleteUser \
  -H "Authorization: Bearer <token)" \
  -H "Content-Type: application/json" \
  -d '{"id":"2"}'
```

- 上传头像

```bash
curl -X POST http://localhost:8090/api/upload-avatar \
  -H "Authorization: Bearer <token>" \
  -F "avatar=@/path/to/avatar.png"
```

**注意事项**

- Token 过期后需重新登录；当角色变更时，后端会在响应头返回 New-Token
- 生产环境请将 JWT 密钥与数据库连接迁移到环境变量，并避免硬编码
- 如果使用 Docker Compose，请统一端口与数据库配置以匹配应用代码当前设置
