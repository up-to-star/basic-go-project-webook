//go:build wireinject

package main

import (
	repository2 "github.com/basic-go-project-webook/webook/interactive/repository"
	cache2 "github.com/basic-go-project-webook/webook/interactive/repository/cache"
	dao2 "github.com/basic-go-project-webook/webook/interactive/repository/dao"
	service2 "github.com/basic-go-project-webook/webook/interactive/service"
	"github.com/basic-go-project-webook/webook/internal/repository"
	"github.com/basic-go-project-webook/webook/internal/repository/article"
	"github.com/basic-go-project-webook/webook/internal/repository/cache"
	"github.com/basic-go-project-webook/webook/internal/repository/dao"
	article2 "github.com/basic-go-project-webook/webook/internal/repository/dao/article"
	"github.com/basic-go-project-webook/webook/internal/service"
	"github.com/basic-go-project-webook/webook/internal/web"
	ijwt "github.com/basic-go-project-webook/webook/internal/web/jwt"
	"github.com/basic-go-project-webook/webook/ioc"
	"github.com/google/wire"
)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	cache.NewRankingLocalCache,
	repository.NewOnlyCachedRankingRepository,
	service.NewBatchRankingService,
)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
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

		// cache 部分
		cache.NewUserCache, cache.NewCodeCache,
		cache.NewRedisArticleCache,

		interactiveSvcSet,
		// ranking
		rankingSvcSet,

		ioc.InitJobs,
		ioc.InitRankingJob,

		// repository
		repository.NewUserRepository,
		repository.NewCodeRepository,
		article.NewArticleRepository,

		ioc.InitInteractiveReadEventConsumer,
		ioc.InitConsumers,

		// service 部分
		ioc.InitSMSService,
		service.NewUserService,
		service.NewCodeService,
		ioc.InitOAuth2WechatService,
		service.NewArticleService,

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
