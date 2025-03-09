package main

import (
	"goWebExample/internal/configs"
)

func main() {

	appConfig := configs.ReadConfig(configs.ConfigPath)
	app := InitializeApp(appConfig)
	app.RunServer()

}
