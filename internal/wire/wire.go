package wire

import (
	"github.com/google/wire"

	"goWebExample/api/rest/handlers"
	"goWebExample/api/rest/handlers/datacenter"
	restuser "goWebExample/api/rest/handlers/user"
	"goWebExample/internal/repository/user"
	"goWebExample/internal/service/datacenter_service"
	"goWebExample/internal/service/user_service"
)

// BusinessSet 包含所有业务相关的依赖
var BusinessSet = wire.NewSet(
	// User 模块
	user.NewUserRepository,
	user_service.NewUserService,
	restuser.NewUserHandler,

	// DataCenter 模块
	datacenter_service.NewMockDataCenter,
	datacenter.NewDataCenterHandler,
	wire.Bind(new(handlers.Handler), new(*datacenter.DataCenterHandler)),
)
