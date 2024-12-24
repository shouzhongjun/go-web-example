package rest

import (
	"github.com/gin-gonic/gin"
	"goWebExample/internal/service/user_service"
)

// UserApi 提供用户相关的 API
type UserApi struct {
	service *user_service.UserService
}

// NewUserApi 创建 UserApi 实例
func NewUserApi(service *user_service.UserService) *UserApi {
	return &UserApi{service: service}
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
