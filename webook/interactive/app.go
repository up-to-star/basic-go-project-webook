package main

import (
	"github.com/basic-go-project-webook/webook/pkg/ginx"
	"github.com/basic-go-project-webook/webook/pkg/grpcx"
	"github.com/basic-go-project-webook/webook/pkg/kafkax"
)

// App 存放所有需要main函数启动、关闭的服务
type App struct {
	server      *grpcx.Server
	consumers   []kafkax.Consumer
	adminServer *ginx.Server
}
