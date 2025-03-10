package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"goWebExample/api/rest/response"
	"goWebExample/internal/service/user"
)

// UserHandler 处理用户相关的HTTP请求
type UserHandler struct {
	userService *user.UserService
}

// NewUserHandler 创建一个新的用户处理器
func NewUserHandler(userService *user.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUserDetail 获取用户详情
func (h *UserHandler) GetUserDetail(c *gin.Context) {
	userId := c.Param("userId")

	user, err := h.userService.GetUserDetail(userId)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Fail(http.StatusNotFound, "用户不存在"))
		return
	}

	c.JSON(http.StatusOK, response.Success(user))
}

// 其他用户相关的处理方法

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	// 实现创建用户的逻辑
	c.JSON(http.StatusOK, response.SuccessWithMessage("创建用户功能待实现", nil))
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// 实现更新用户的逻辑
	userId := c.Param("userId")
	c.JSON(http.StatusOK, response.SuccessWithMessage("更新用户功能待实现", gin.H{
		"userId": userId,
	}))
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 实现删除用户的逻辑
	userId := c.Param("userId")
	c.JSON(http.StatusOK, response.SuccessWithMessage("删除用户功能待实现", gin.H{
		"userId": userId,
	}))
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 实现获取用户列表的逻辑
	c.JSON(http.StatusOK, response.SuccessWithMessage("获取用户列表功能待实现", nil))
}

// RegisterRoutes 注册用户相关路由
// 实现 handlers.Handler 接口
func (h *UserHandler) RegisterRoutes(apiGroup *gin.RouterGroup) {
	userGroup := apiGroup.Group("/users")
	{
		// 基本用户操作
		userGroup.GET("/:userId", h.GetUserDetail)
		userGroup.POST("", h.CreateUser)
		userGroup.PUT("/:userId", h.UpdateUser)
		userGroup.DELETE("/:userId", h.DeleteUser)
		userGroup.GET("", h.ListUsers)
	}
}
