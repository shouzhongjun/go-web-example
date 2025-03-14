package main

import (
	"log"

	"goWebExample/internal/configs"
)

func main() {
	// 加载配置
	config := configs.ReadConfig(configs.ConfigPath)
	if config == nil {
		log.Fatal("加载配置失败")
	}

	// 初始化应用程序
	application, err := InitializeApp(config)
	if err != nil {
		log.Fatalf("初始化应用程序失败: %v", err)
	}

	// 运行应用程序
	if err := application.Run(); err != nil {
		log.Fatalf("运行应用程序失败: %v", err)
	}
}
