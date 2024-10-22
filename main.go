package main

import (
	"context"
	"distributed-object-storage/config"
	"distributed-object-storage/controller"
	_ "distributed-object-storage/docs"
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/pkg/log"
	"distributed-object-storage/redis"
	"distributed-object-storage/syncer"
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
	cfg    *config.Config
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
	app.server.Use(cors.Default())
	//app.server.Use(middwares.AuthMiddleware())
	initApp(app)
	port := 3002
	log.Infof("Server starting on port %v", port)
	err := app.server.Run(fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return
	}
}

func initApp(app *Commands) {
	dos := dao.Init()
	if err := redis.Init(); err != nil {
		log.Errorf("Redis can not init %v", err)
	}
	metaDataController := controller.NewMetadataNodeController(dos)
	storageController := controller.NewStorageNodeController(dos)
	authController := controller.NewAuthController(dos)
	app.server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	metaDataController.RegisterRouter(app.server)
	storageController.RegisterRouter(app.server)
	authController.RegisterRouter(app.server)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		syncer.Init(ctx)
	}()
}
