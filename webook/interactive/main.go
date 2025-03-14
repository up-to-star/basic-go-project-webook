package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	initViper()
	initZap()
	initPrometheus()

	app := InitAPP()
	for _, c := range app.consumers {
		c.Start()
	}

	go func() {
		err1 := app.adminServer.Start()
		if err1 != nil {
			panic(err1)
		}
	}()
	server := app.server
	err := server.Serve()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	viper.SetDefault("db.mysql.dsn",
		"root:root@tcp(localhost:13316)/webook?charset=utf8mb4&parseTime=True&loc=Local")
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("viper 启动失败: %s \n", err))
	}
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8082", nil)
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
