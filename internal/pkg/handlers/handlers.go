package handlers

import (
	"goWebExample/api/rest/handlers"
	"goWebExample/api/rest/handlers/datacenter"
)

// Handlers 包含所有HTTP处理器
type Handlers struct {
	User       *handlers.UserHandler
	DataCenter *datacenter.DataCenterHandler
}
