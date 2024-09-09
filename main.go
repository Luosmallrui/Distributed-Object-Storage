package main

import (
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli"
)

type Commands struct {
	app    *cli.App
	server *gin.Engine
}

func NewCommands() *Commands {
	cmd := &Commands{
		app:    cli.NewApp(),
		server: gin.Default(), // 初始化 server
	}

	return cmd
}

func main() {
	app := NewCommands()
	//logs.InitLogger()
	//app.server.Use(middwares.Cors())
	app.server.Use(gin.Logger())
	//app.server.Use(middwares.AuthMiddleware())
	initApp(app)
	err := app.server.Run("0.0.0.0:3003")
	if err != nil {
		return
	}
}

func initApp(app *Commands) {
	//dos := dao.Init()
	//if err := redis.Init(); err != nil {
	//	logs.Logger.Errorf("Redis can not init %v", err)
	//}
}
