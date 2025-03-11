# 分层架构示例代码

本文档展示了基于分层架构的代码示例，包含 API、Service、Repository 层的实现。

## 目录结构
```
.
├── api
│   └── rest
│       ├── handlers
│       │   └── user
│       │       └── user_handler.go
│       └── response
│           └── response.go
├── internal
│   ├── pkg
│   │   ├── handlers
│   │   │   ├── handler.go
│   │   │   └── registry.go
│   │   └── module
│   │       └── module.go
│   ├── repository
│   │   └── user
│   │       └── repository.go
│   └── service
│       ├── registry.go
│       └── user
│           └── service.go
```

## API 层示例

### `api/rest/response/response.go`
```go
package response

// Response 通用API响应结构
type Response struct {
    Code    int         `json:"code"`    // 状态码
    Message string      `json:"message"` // 消息
    Data    interface{} `json:"data"`    // 数据
}

// Success 返回成功响应
func Success(data interface{}) Response {
    return Response{
        Code:    200,
        Message: "success",
        Data:    data,
    }
}

// Fail 返回失败响应
func Fail(code int, message string) Response {
    return Response{
        Code:    code,
        Message: message,
        Data:    nil,
    }
}
```

### `api/rest/handlers/user/user_handler.go`
```go
package user

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "goWebExample/api/rest/response"
    "goWebExample/internal/pkg/handlers"
    "goWebExample/internal/service"
    "goWebExample/internal/service/user"
)

// UserHandler 处理用户相关的HTTP请求
type UserHandler struct {
    logger *zap.Logger
}

// NewUserHandler 创建用户处理器
func NewUserHandler(logger *zap.Logger) *UserHandler {
    return &UserHandler{
        logger: logger,
    }
}

// GetRouteGroup 获取路由组
func (h *UserHandler) GetRouteGroup() handlers.RouteGroup {
    return handlers.API
}

// RegisterRoutes 注册路由
func (h *UserHandler) RegisterRoutes(group *gin.RouterGroup) {
    userGroup := group.Group("/users")
    {
        userGroup.GET("/:userId", h.GetUserDetail)
        userGroup.POST("", h.CreateUser)
        userGroup.PUT("/:userId", h.UpdateUser)
        userGroup.DELETE("/:userId", h.DeleteUser)
        userGroup.GET("", h.ListUsers)
    }
}

// GetUserDetail 获取用户详情
func (h *UserHandler) GetUserDetail(c *gin.Context) {
    // 从服务注册器获取服务
    srv, ok := service.GetRegistry().Get(user.ServiceName).(*user.UserService)
    if !ok || srv == nil {
        h.logger.Error("user service not initialized")
        c.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "用户服务未初始化"))
        return
    }

    userId := c.Param("userId")
    user, err := srv.GetUserDetail(userId)
    if err != nil {
        h.logger.Error("failed to get user detail", zap.Error(err))
        c.JSON(http.StatusNotFound, response.Fail(http.StatusNotFound, "用户不存在"))
        return
    }

    c.JSON(http.StatusOK, response.Success(user))
}
```

## Service 层示例

### `internal/service/user/service.go`
```go
package user

import (
    "goWebExample/internal/repository/user"
)

const ServiceName = "user"

// UserService 用户服务
type UserService struct {
    repo user.Repository
}

// NewUserService 创建用户服务
func NewUserService(repo user.Repository) *UserService {
    return &UserService{
        repo: repo,
    }
}

// User 用户模型
type User struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    Status   int    `json:"status"`
}

// GetUserDetail 获取用户详情
func (s *UserService) GetUserDetail(userID string) (*User, error) {
    user, err := s.repo.GetByID(userID)
    if err != nil {
        return nil, err
    }

    return &User{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
        Phone:    user.Phone,
        Status:   user.Status,
    }, nil
}
```

## Repository 层示例

### `internal/repository/user/repository.go`
```go
package user

import (
    "database/sql"
    "errors"
)

// Repository 用户仓库接口
type Repository interface {
    GetByID(id string) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id string) error
    List(offset, limit int) ([]*User, int, error)
}

// User 用户数据模型
type User struct {
    ID       string
    Username string
    Email    string
    Phone    string
    Status   int
}

// SQLRepository MySQL实现的用户仓库
type SQLRepository struct {
    db *sql.DB
}

// NewSQLRepository 创建MySQL用户仓库
func NewSQLRepository(db *sql.DB) Repository {
    return &SQLRepository{
        db: db,
    }
}

// GetByID 根据ID获取用户
func (r *SQLRepository) GetByID(id string) (*User, error) {
    var user User
    err := r.db.QueryRow(
        "SELECT id, username, email, phone, status FROM users WHERE id = ?",
        id,
    ).Scan(&user.ID, &user.Username, &user.Email, &user.Phone, &user.Status)

    if err == sql.ErrNoRows {
        return nil, errors.New("user not found")
    }
    if err != nil {
        return nil, err
    }

    return &user, nil
}

// Create 创建用户
func (r *SQLRepository) Create(user *User) error {
    result, err := r.db.Exec(
        "INSERT INTO users (id, username, email, phone, status) VALUES (?, ?, ?, ?, ?)",
        user.ID, user.Username, user.Email, user.Phone, user.Status,
    )
    if err != nil {
        return err
    }

    affected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if affected == 0 {
        return errors.New("failed to create user")
    }

    return nil
}
```

## 模块注册示例

### `internal/pkg/module/module.go`
```go
package module

import (
    "goWebExample/internal/infra/di/container"
    "goWebExample/internal/pkg/handlers"
    "goWebExample/internal/service"
    "goWebExample/internal/service/user"
    userRepo "goWebExample/internal/repository/user"

    "go.uber.org/zap"
)

func init() {
    // 注册用户模块
    GetRegistry().Register(NewBaseModule(
        "user",
        // 服务创建函数
        func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}) {
            if container != nil && container.DBConnector != nil {
                userRepository := userRepo.NewSQLRepository(container.DBConnector.DB)
                userSvc := user.NewUserService(userRepository)
                return user.ServiceName, userSvc
            }
            logger.Error("无法初始化用户服务：数据库连接器未初始化")
            return "", nil
        },
        // 处理器创建函数
        func(logger *zap.Logger) handlers.Handler {
            return NewUserHandler(logger)
        },
    ))
}
```

## 使用示例

### `cmd/app/main.go`
```go
package main

import (
    "context"
    "log"

    "goWebExample/internal/app"
    "goWebExample/internal/configs"
    "goWebExample/pkg/zap"
)

func main() {
    // 加载配置
    config := configs.LoadConfig()

    // 初始化日志
    logger := zaplogger.NewZap()

    // 创建应用实例
    application, err := InitializeApp(config, logger)
    if err != nil {
        log.Fatal(err)
    }

    // 运行应用
    if err := application.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

## 最佳实践

### 分层职责
1. API 层
   - 处理 HTTP 请求和响应
   - 参数验证和转换
   - 错误处理和响应格式化
   - 路由注册

2. Service 层
   - 业务逻辑处理
   - 事务管理
   - 数据转换和组装
   - 调用其他服务

3. Repository 层
   - 数据访问和持久化
   - SQL 查询和执行
   - 数据模型映射
   - 缓存管理

### 依赖注入
1. 使用 Wire 进行依赖注入
2. 面向接口编程
3. 组件生命周期管理
4. 配置的集中管理

### 错误处理
1. 定义领域错误类型
2. 错误包装和转换
3. 统一的错误响应格式
4. 错误日志记录

### 代码组织
1. 按功能模块划分目录
2. 遵循依赖倒置原则
3. 使用接口定义契约
4. 保持单一职责原则 