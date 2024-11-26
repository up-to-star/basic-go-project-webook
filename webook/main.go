package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	//db := initDB()
	//redisClient := rds.NewClient(&rds.Options{
	//	Addr: config.Config.Redis.Addr,
	//})
	//u := initUser(db, redisClient)
	//server := initWebServer(redisClient)
	//u.RegisterRoutes(server)
	initViper()
	//initViperRemote()
	initZap()
	server := InitWebServer()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello world")
	})
	_ = server.Run(":8080")
}

func initZap() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func initViper() {
	viper.SetDefault("db.mysql.dsn",
		"root:root@tcp(localhost:13316)/webook?charset=utf8mb4&parseTime=True&loc=Local")
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./webook/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("viper 启动失败: %s \n", err))
	}
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3",
		"http://localhost:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}
