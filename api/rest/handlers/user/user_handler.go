package user

import (
	"goWebExample/api/rest/handlers/user/request"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"goWebExample/api/rest/response"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/module"
	userRepo "goWebExample/internal/repository/user"
	"goWebExample/internal/service"
	"goWebExample/internal/service/user"
)

func init() {
	// 注册模块
	module.GetRegistry().Register(module.NewBaseModule(
		"user",
		// 服务创建函数
		func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}) {
			if container != nil && container.DBConnector != nil {
				userRepository := userRepo.NewUserRepository(container.DBConnector)
				userSvc := user.NewUserService(userRepository, logger)
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

// UserHandler 处理用户相关的HTTP请求
type UserHandler struct {
	logger *zap.Logger
}

// NewUserHandler 创建一个新的用户处理器
func NewUserHandler(logger *zap.Logger) *UserHandler {
	return &UserHandler{
		logger: logger,
	}
}

// GetRouteGroup 获取路由组
func (h *UserHandler) GetRouteGroup() handlers.RouteGroup {
	return handlers.API
}

// GetUserDetail 获取用户详情
func (h *UserHandler) GetUserDetail(c *gin.Context) {
	// 从服务注册器获取服务
	srv, ok := service.GetRegistry().Get(user.ServiceName).(*user.UserService)
	if !ok || srv == nil {
		h.logger.Error("userDetail service not initialized")
		c.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "用户服务未初始化"))
		return
	}

	userId := c.Param("userId")
	userDetail, err := srv.GetUserDetail(userId)
	if err != nil {
		h.logger.Error("failed to get userDetail detail", zap.Error(err))
		c.JSON(http.StatusNotFound, response.Fail(http.StatusNotFound, "用户不存在"))
		return
	}

	response.SuccessWithData(c, userDetail)
	return
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	// 从服务注册器获取服务
	srv, ok := service.GetRegistry().Get(user.ServiceName).(*user.UserService)
	if !ok || srv == nil {
		h.logger.Error("user service not initialized")
		c.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "用户服务未初始化"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage("创建用户功能待实现", nil))
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// 从服务注册器获取服务
	srv, ok := service.GetRegistry().Get(user.ServiceName).(*user.UserService)
	if !ok || srv == nil {
		h.logger.Error("user service not initialized")
		c.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "用户服务未初始化"))
		return
	}

	userId := c.Param("userId")
	c.JSON(http.StatusOK, response.SuccessWithMessage("更新用户功能待实现", gin.H{
		"userId": userId,
	}))
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 从服务注册器获取服务
	srv, ok := service.GetRegistry().Get(user.ServiceName).(*user.UserService)
	if !ok || srv == nil {
		h.logger.Error("user service not initialized")
		c.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "用户服务未初始化"))
		return
	}

	userId := c.Param("userId")
	c.JSON(http.StatusOK, response.SuccessWithMessage("删除用户功能待实现", gin.H{
		"userId": userId,
	}))
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 从服务注册器获取服务
	srv, ok := service.GetRegistry().Get(user.ServiceName).(*user.UserService)
	if !ok || srv == nil {
		h.logger.Error("user service not initialized")
		c.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "用户服务未初始化"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage("获取用户列表功能待实现", nil))
}

func (h *UserHandler) LoginHandler(ctx *gin.Context) {
	// 从服务注册器获取服务
	srv, ok := service.GetRegistry().Get(user.ServiceName).(*user.UserService)
	if !ok || srv == nil {
		h.logger.Error("user service not initialized")
		ctx.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, "用户服务未初始化"))
		return
	}
	var req *request.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Fail(http.StatusBadRequest, "参数错误"))
		return
	}
	// 获取客户端IP
	clientIP := ctx.ClientIP()
	users, err := srv.Login(req.Username, req.Password, clientIP)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Fail(http.StatusInternalServerError, err.Error()))
		return
	}
	response.SuccessWithData(ctx, users)
	return
}

// RegisterRoutes 注册用户相关路由
// 实现 handlers.Handler 接口
func (h *UserHandler) RegisterRoutes(apiGroup *gin.RouterGroup) {
	if h == nil {
		panic("UserHandler is nil when registering routes")
	}

	userGroup := apiGroup.Group("/users")
	{
		// 基本用户操作
		userGroup.GET("/:userId", h.GetUserDetail)
		userGroup.POST("", h.CreateUser)
		userGroup.PUT("/:userId", h.UpdateUser)
		userGroup.DELETE("/:userId", h.DeleteUser)
		userGroup.GET("", h.ListUsers)
		userGroup.POST("/login", h.LoginHandler)
	}
}
