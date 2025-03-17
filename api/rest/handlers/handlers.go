package handlers

import (
	// 在这里导入所有的 handlers
	_ "goWebExample/api/rest/handlers/datacenter"
	_ "goWebExample/api/rest/handlers/ly_stop"
	_ "goWebExample/api/rest/handlers/stream"
	_ "goWebExample/api/rest/handlers/user"
)

// Register 注册所有处理器
// 这个函数是为了确保所有 handler 包的 init() 函数都被执行
func Register() {
	// 空函数，仅用于触发 init
}
