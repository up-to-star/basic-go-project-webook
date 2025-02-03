//go:build wireinject

package main

import (
	"basic-project/webook/internal/repository"
	"basic-project/webook/internal/repository/article"
	"basic-project/webook/internal/repository/cache"
	"basic-project/webook/internal/repository/dao"
	article2 "basic-project/webook/internal/repository/dao/article"
	"basic-project/webook/internal/service"
	"basic-project/webook/internal/web"
	ijwt "basic-project/webook/internal/web/jwt"
	"basic-project/webook/ioc"
	"github.com/google/wire"
)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewOnlyCachedRankingRepository,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitProducer,
		//ioc.InitMongoDB,
		//ioc.InitSnowFlakeNode,
		ioc.InitRlockClient,

		// dao 部分
		dao.NewUserDAO,
		article2.NewArticleDAO,
		//article2.NewMongoDBArticleDAO,
		dao.NewGORMInteractiveDAO,

		// cache 部分
		cache.NewUserCache, cache.NewCodeCache,
		cache.NewRedisArticleCache,
		cache.NewInteractiveRedisCache,

		// ranking
		rankingSvcSet,

		ioc.InitJobs,
		ioc.InitRankingJob,

		// repository
		repository.NewUserRepository, repository.NewCodeRepository,
		article.NewArticleRepository,
		repository.NewCachedInteractiveRepository,
		ioc.InitInteractiveReadEventConsumer,
		ioc.InitConsumers,

		// service 部分
		ioc.InitSMSService,
		service.NewUserService,
		service.NewCodeService,
		ioc.InitOAuth2WechatService,
		service.NewArticleService,
		service.NewInteractiveService,

		// handler 部分
		ijwt.NewRedisJwtHandler,
		web.NewUserHandle,
		web.NewArticleHandle,
		web.NewOAuth2WechatHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebserver,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
