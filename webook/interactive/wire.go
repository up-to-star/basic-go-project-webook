//go:build wireinject

package main

import (
	"github.com/basic-go-project-webook/webook/interactive/grpc"
	"github.com/basic-go-project-webook/webook/interactive/ioc"
	"github.com/basic-go-project-webook/webook/interactive/repository"
	"github.com/basic-go-project-webook/webook/interactive/repository/cache"
	"github.com/basic-go-project-webook/webook/interactive/repository/dao"
	"github.com/basic-go-project-webook/webook/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	ioc.InitSrcDB,
	ioc.InitDstDB,
	ioc.InitDoubleWritePool,
	ioc.InitBizDB,
	ioc.InitRedis,
)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitAPP() *App {
	wire.Build(
		interactiveSvcSet,
		thirdPartySet,
		grpc.NewInteractiveServiceServer,
		ioc.InitInconsistentProducer,
		ioc.InitInteractiveReadEventConsumer,
		ioc.InitFixerConsumer,
		ioc.InitConsumers,
		ioc.InitGRPCXServer,
		ioc.InitGinxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
