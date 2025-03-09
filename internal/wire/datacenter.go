package wire

import (
	"goWebExample/api/rest/handlers"
	"goWebExample/internal/service/datacenter_service"

	"github.com/google/wire"
)

// DataCenterModule 数据中心模块相关依赖
var DataCenterModule = wire.NewSet(
	// Service
	datacenter_service.NewMockDataCenter,

	// Handler
	handlers.NewDataCenterHandler,
)
