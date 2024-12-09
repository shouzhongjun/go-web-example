package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"goWebExample/internal/service/user_service"
	"gorm.io/gorm"
)

type App struct {
	DB          *gorm.DB
	Redis       *redis.Client
	Kafka       *kafka.Writer
	UserService *user_service.UserService
}

// NewApp 是 *App 的构造函数，接收所有依赖项作为参数
func NewApp(db *gorm.DB, redis *redis.Client, kafka *kafka.Writer, userService *user_service.UserService) *App {
	return &App{
		DB:          db,
		Redis:       redis,
		Kafka:       kafka,
		UserService: userService,
	}
}

// Run 是应用启动的逻辑
func (a *App) Run() {
	// 启动应用的逻辑，比如启动 HTTP 服务器、任务队列等
	art := `
           __        __   _     _____                           _      
   __ _  __\ \      / /__| |__ | ____|_  ____ _ _ __ ___  _ __ | | ___ 
  / _` + "`" + ` |/ _ \ \ /\ / / _ \ '_ \|  _| \ \/ / _` + "`" + ` | '_ ` + "`" + ` _ \| '_ \| |/ _ \
 | (_| | (_) \ V  V /  __/ |_) | |___ >  < (_| | | | | | | |_) | |  __/
  \__, |\___/ \_/\_/ \___|_.__/|_____/_/\_\__,_|_| |_| |_| .__/|_|\___|
  |___/                                                  |_|           
`
	fmt.Print(art)
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run(":8080")
}
