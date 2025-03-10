package wire

import (
	"github.com/google/wire"

	"goWebExample/api/rest/handlers"
	restdc "goWebExample/api/rest/handlers/datacenter"
	restuser "goWebExample/api/rest/handlers/user"
	"goWebExample/internal/repository/user"
	dc "goWebExample/internal/service/datacenter"
	userservice "goWebExample/internal/service/user"
)

// BusinessSet 包含所有业务相关的依赖
var BusinessSet = wire.NewSet(
	// User 模块
	user.NewUserRepository,
	userservice.NewUserService,
	restuser.NewUserHandler,

	// DataCenter 模块
	dc.NewMockDataCenter,
	restdc.NewDataCenterHandler,
	wire.Bind(new(handlers.Handler), new(*restdc.DataCenterHandler)),
)
