package wire

import (
	"github.com/google/wire"
)

// AllModules 包含所有业务模块的依赖
var AllModules = wire.NewSet(
	UserModule,
	DataCenterModule,
	// 未来可以在这里添加更多模块
)
