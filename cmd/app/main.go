package main

import (
	"goWebExample/internal/app" // 引入 internal/app 包
	"goWebExample/internal/service/user_service"
)

func main() {
	// 初始化应用程序
	application, _ := app.InitApp()
	user_service.Create("123", "123")
	// 启动应用
	application.Run()

}
