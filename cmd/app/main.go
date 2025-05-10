package main

import (
	"flag"
	"fmt"
	_ "goWebExample/docs/swagger" // 导入 swagger docs
	"goWebExample/internal/configs"
	"goWebExample/internal/version"
	"log"
)

func init() {
	// 避免重复调用获取版本信息
	version.Version, version.BuildTime, version.CommitSHA = version.GetRuntimeVersionInfo()
}

// @title GoWebExample API
// @version 1.0
// @description This is a sample server for GoWebExample.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api
// @schemes   http https

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	flag.Parse()

	// 直接使用已经在 init 函数中设置好的版本信息
	fmt.Printf("Version   : %s\n", version.Version)
	fmt.Printf("Build Time: %s\n", version.BuildTime)
	fmt.Printf("Git SHA   : %s\n", version.CommitSHA)
	fmt.Println("----------------------------------------")

	// 读取配置文件并进行错误处理
	config := configs.ReadConfig(configs.ConfigPath)
	// 初始化并启动应用
	app, err := InitializeApp(config)
	if err != nil {
		log.Fatalf("初始化应用失败: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("启动应用失败: %v", err)
	}
}
