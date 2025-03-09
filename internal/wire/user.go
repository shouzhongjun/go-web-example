package wire

import (
	"goWebExample/api/rest/handlers"
	"goWebExample/internal/repository/user"
	"goWebExample/internal/service/user_service"

	"github.com/google/wire"
)

// UserModule 用户模块相关依赖
var UserModule = wire.NewSet(
	// Repository
	user.NewUserRepository,

	// Service
	user_service.NewUserService,

	// Handler
	handlers.NewUserHandler,
)
