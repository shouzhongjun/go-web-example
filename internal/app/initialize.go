package app

// InitApp 初始化应用，返回 App 实例或错误
func InitApp() *App {
	// 调用 Wire 生成的初始化函数来创建 App 实例
	app, err := InitializeApp()
	if err != nil {
		panic(err)
	}
	return app
}
