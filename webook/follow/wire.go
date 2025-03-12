//go:build wireinject

package main

import (
	"github.com/basic-go-project-webook/webook/follow/grpc"
	"github.com/basic-go-project-webook/webook/follow/ioc"
	"github.com/basic-go-project-webook/webook/follow/repository"
	"github.com/basic-go-project-webook/webook/follow/repository/cache"
	"github.com/basic-go-project-webook/webook/follow/repository/dao"
	"github.com/basic-go-project-webook/webook/follow/service"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
)

var serviceProvider = wire.NewSet(
	dao.NewGORMFollowDAO,
	cache.NewRedisFollowCache,
	repository.NewFollowRepository,
	service.NewFollowService,
	grpc.NewFollowServiceServer,
)

func InitApp() *App {
	wire.Build(
		thirdProvider,
		serviceProvider,
		ioc.InitGRPCXServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
