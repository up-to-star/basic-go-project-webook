package main

import (
	"context"
	"fmt"
	"github.com/basic-go-project-webook/webook/ioc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {

	initViper()

	initZap()

	tpCancel := ioc.InitOTEL()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		tpCancel(ctx)
	}()

	initPrometheus()

	app := InitWebServer()
	for _, consumer := range app.consumers {
		consumer.Start()
	}
	// 启动定时任务
	app.cron.Start()
	defer func() {
		// 等待定时任务退出
		<-app.cron.Stop().Done()
	}()
	server := app.web
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			panic(err)
		}
	}()
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
