package main

import (
	"goWebExample/internal/configs"
)

func main() {

	config := configs.ReadConfig(configs.ConfigPath)
	wireApp := WireApp(config)
	wireApp.RunServer()

}
