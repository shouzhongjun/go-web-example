package service

import (
	"github.com/google/wire"
	"goWebExample/internal/repository/user"
	"goWebExample/internal/service/user_service"
)

// ServicesProvider 是服务层业务逻辑的 Wire provider
// userProviderSet 是用户相关服务的 Wire provider
var ServicesProvider = wire.NewSet(
	userProviderSet,
)

// userProviderSet 是服务层用户相关业务逻辑的 Wire provider
var userProviderSet = wire.NewSet(
	user_service.NewUserService,
	user.NewUserRepository,
)
