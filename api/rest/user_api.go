package rest

import (
	"goWebExample/internal/service/user_service"

	"github.com/gin-gonic/gin"
)

// UserApi 是用户相关API的处理结构体
type UserApi struct {
	service *user_service.UserService
}

// NewUserApi 创建 UserApi 实例
func NewUserApi(service *user_service.UserService) *UserApi {
	return &UserApi{
		service: service,
	}
}

// GetUserDetail 获取用户详情
func (u *UserApi) GetUserDetail(ctx *gin.Context) {
	userId := ctx.Param("userId")
	detail, err := u.service.GetUserDetail(userId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, detail)

}
