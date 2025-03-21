package users

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"goWebExample/api/protobuf/users/pb"
	"goWebExample/internal/infra/di/container"
	"goWebExample/internal/pkg/handlers"
	"goWebExample/internal/pkg/module"
	userRepo "goWebExample/internal/repository/user"
	"goWebExample/internal/service"
	"goWebExample/internal/service/user"
)

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

// RegisterRoutes 注册路由
func (h *UserHandler) RegisterRoutes(apiGroup *gin.RouterGroup) {
	apiGroup.POST("/users/proto/login", h.Login)
}

// Login 处理登录请求
func (h *UserHandler) Login(ctx *gin.Context) {
	var req pb.LoginRequest

	// 根据 Content-Type 选择解析方式
	contentType := ctx.GetHeader("Content-Type")
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		h.logger.Error("读取请求数据失败", zap.Error(err))
		ctx.Status(http.StatusBadRequest)
		return
	}

	switch {
	case strings.Contains(contentType, "application/json"):
		// 解析 JSON
		if err := protojson.Unmarshal(body, &req); err != nil {
			h.logger.Error("解析JSON数据失败", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的JSON数据"})
			return
		}
	case strings.Contains(contentType, "application/x-protobuf"):
		// 解析 protobuf
		if err := proto.Unmarshal(body, &req); err != nil {
			h.logger.Error("解析protobuf数据失败", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的protobuf数据"})
			return
		}
	case strings.Contains(contentType, "text/plain"):
		// 尝试解析 base64 编码的 protobuf 数据
		decoded, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			h.logger.Error("解析base64数据失败", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的base64数据"})
			return
		}
		if err := proto.Unmarshal(decoded, &req); err != nil {
			h.logger.Error("解析protobuf数据失败", zap.Error(err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的protobuf数据"})
			return
		}
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "不支持的Content-Type"})
		return
	}

	// 获取用户服务实例
	srv, ok := service.GetRegistry().Get(user.ServiceName).(*user.UserService)
	if !ok || srv == nil {
		h.logger.Error("用户服务未初始化")
		resp := &pb.LoginResponse{
			Error: "用户服务未初始化",
		}

		// 根据请求的 Content-Type 返回相应格式的响应
		if strings.Contains(contentType, "application/json") {
			ctx.JSON(http.StatusInternalServerError, resp)
		} else {
			data, _ := proto.Marshal(resp)
			ctx.Data(http.StatusInternalServerError, "application/x-protobuf", data)
		}
		return
	}

	// 处理登录请求
	result, err := srv.Login(req.Username, req.Password, ctx.ClientIP())
	if err != nil {
		h.logger.Error("登录失败", zap.Error(err))
		resp := &pb.LoginResponse{
			Error: err.Error(),
		}

		// 根据请求的 Content-Type 返回相应格式的响应
		if strings.Contains(contentType, "application/json") {
			ctx.JSON(http.StatusUnauthorized, resp)
		} else {
			data, _ := proto.Marshal(resp)
			ctx.Data(http.StatusUnauthorized, "application/x-protobuf", data)
		}
		return
	}

	// 构造成功响应
	resp := &pb.LoginResponse{
		Token:    result.AccessToken,
		Nickname: result.User.Nickname,
		Email:    result.User.Email,
	}

	// 根据请求的 Content-Type 返回相应格式的响应
	if strings.Contains(contentType, "application/json") {
		ctx.JSON(http.StatusOK, resp)
	} else {
		data, err := proto.Marshal(resp)
		if err != nil {
			h.logger.Error("序列化响应失败", zap.Error(err))
			ctx.Status(http.StatusInternalServerError)
			return
		}
		ctx.Header("Content-Type", "application/x-protobuf")
		ctx.Data(http.StatusOK, "application/x-protobuf", data)
	}
}

func init() {
	// 注册模块
	module.GetRegistry().Register(module.NewBaseModule(
		"user",
		// 服务创建函数
		func(logger *zap.Logger, container *container.ServiceContainer) (string, interface{}) {
			if container != nil && container.DBConnector != nil {
				userRepository := userRepo.NewUserRepository(container.DBConnector)
				jwtManager := container.GetJWTManager()
				if jwtManager == nil {
					logger.Error("无法初始化用户服务：JWT管理器未初始化")
					return "", nil
				}
				userSvc := user.NewUserService(userRepository, logger, jwtManager)
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
