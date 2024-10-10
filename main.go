package main

import (
	"distributed-object-storage/config"
	"distributed-object-storage/controller"
	_ "distributed-object-storage/docs"
	"distributed-object-storage/pkg/db/dao"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"github.com/urfave/cli"
	"os"
)

type Commands struct {
	app    *cli.App
	server *gin.Engine
	cfg    *config.Config // 框架基础配置
}

func (cmd *Commands) GetConfig() *config.Config {
	return cmd.cfg
}

func NewCommands() *Commands {
	cmd := &Commands{
		app:    cli.NewApp(),
		server: gin.Default(), // 初始化 server
	}
	cmd.app.Action = func(cli *cli.Context) {
		cmd.cfg = config.InitConfig(cli)
	}
	if err := cmd.app.Run(os.Args); err != nil {
		fmt.Println(err)
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
	fmt.Println(app.cfg.OssConfig)
	err := app.server.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}

func initApp(app *Commands) {
	dos := dao.Init()
	metaDataController := controller.NewMetadataNodeController(dos)
	storageController := controller.NewStorageNodeController(dos)
	//if err := redis.Init(); err != nil {
	//	logs.Logger.Errorf("Redis can not init %v", err)
	//}

	app.server.Use(cors.Default())
	app.server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	metaDataController.RegisterRouter(app.server)
	storageController.RegisterRouter(app.server)
}
