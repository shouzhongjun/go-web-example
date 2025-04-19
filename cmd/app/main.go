package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	_ "goWebExample/docs/swagger" // 导入 swagger docs
	"goWebExample/internal/configs"
)

// 版本信息，通过 -ldflags 注入
var (
	Version   = "dev"
	BuildTime = "unknown"
	CommitSHA = "unknown"
)

// @title           GoWebExample API
// @version         1.0
// @description     This is a sample server for GoWebExample.
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

// 获取运行时的版本信息
func getRuntimeVersionInfo() (version, buildTime, commitSHA string) {
	// 如果通过 -ldflags 注入了值，直接返回
	if Version != "dev" || BuildTime != "unknown" || CommitSHA != "unknown" {
		return Version, BuildTime, CommitSHA
	}

	// 否则尝试从 git 获取信息
	var err error

	// 获取分支名
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err == nil {
		version = strings.TrimSpace(string(out))
	} else {
		version = "dev"
	}

	// 获取 commit hash
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err = cmd.Output()
	if err == nil {
		commitSHA = strings.TrimSpace(string(out))
	}

	// 使用当前时间（本地时间）
	buildTime = time.Now().Format("2006-01-02 15:04:05")

	return
}

func main() {
	flag.Parse()

	// 获取版本信息
	version, buildTime, commitSHA := getRuntimeVersionInfo()

	// 输出版本信息
	fmt.Printf("Version   : %s\n", version)
	fmt.Printf("Build Time: %s\n", buildTime)
	fmt.Printf("Git SHA   : %s\n", commitSHA)
	fmt.Println("----------------------------------------")

	// 读取配置文件
	config := configs.ReadConfig(configs.ConfigPath)
	if config == nil {
		log.Fatal("读取配置文件失败")
	}

	// 初始化应用
	app, err := InitializeApp(config)
	if err != nil {
		log.Fatal("初始化应用失败:", err)
	}

	// 启动应用
	if err := app.Run(); err != nil {
		log.Fatal("启动应用失败:", err)
	}
}
