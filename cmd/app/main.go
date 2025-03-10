package main

import (
	"flag"
	"fmt"

	"goWebExample/internal/configs"
)

var (
	port = flag.Int("port", 8080, "服务器端口")
)

func main() {
	// 解析命令行参数
	flag.Parse()
	// 加载配置
	config := configs.ReadConfig(configs.ConfigPath)
	if config == nil {
		panic("加载配置失败")
	}

	// 如果命令行指定了端口，则覆盖配置文件中的端口
	if *port != 0 {
		config.Server.Port = *port
	}

	// 创建应用程序
	app, err := InitializeApp(config)
	if err != nil {
		panic(fmt.Sprintf("初始化应用程序失败: %v", err))
	}

	// 运行应用程序
	app.GetHTTPServer().RunServer()
}
