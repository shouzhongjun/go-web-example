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
	// 初始化 version 包中的版本信息
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

	// 获取并输出版本信息
	v, bt, sha := version.GetRuntimeVersionInfo()
	fmt.Printf("Version   : %s\n", v)
	fmt.Printf("Build Time: %s\n", bt)
	fmt.Printf("Git SHA   : %s\n", sha)
	fmt.Println("----------------------------------------")

	// 读取配置文件
	config := configs.ReadConfig(configs.ConfigPath)
	if config == nil {
		log.Fatal("读取配置文件失败")
	}

	// 初始化并启动应用
	app, err := InitializeApp(config)
	if err != nil {
		log.Fatal("初始化应用失败:", err)
	}

	if err := app.Run(); err != nil {
		log.Fatal("启动应用失败:", err)
	}
}
