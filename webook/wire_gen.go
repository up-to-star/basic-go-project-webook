// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	repository2 "github.com/basic-go-project-webook/webook/interactive/repository"
	cache2 "github.com/basic-go-project-webook/webook/interactive/repository/cache"
	dao2 "github.com/basic-go-project-webook/webook/interactive/repository/dao"
	service2 "github.com/basic-go-project-webook/webook/interactive/service"
	"github.com/basic-go-project-webook/webook/internal/repository"
	article2 "github.com/basic-go-project-webook/webook/internal/repository/article"
	"github.com/basic-go-project-webook/webook/internal/repository/cache"
	"github.com/basic-go-project-webook/webook/internal/repository/dao"
	"github.com/basic-go-project-webook/webook/internal/repository/dao/article"
	"github.com/basic-go-project-webook/webook/internal/service"
	"github.com/basic-go-project-webook/webook/internal/web"
	"github.com/basic-go-project-webook/webook/internal/web/jwt"
	"github.com/basic-go-project-webook/webook/ioc"
	"github.com/google/wire"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitWebServer() *App {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJwtHandler(cmdable)
	v := ioc.InitGinMiddlewares(cmdable, handler)
	db := ioc.InitDB()
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandle := web.NewUserHandle(userService, codeService, cmdable, handler)
	wechatService := ioc.InitOAuth2WechatService()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	articleDAO := article.NewArticleDAO(db)
	articleCache := cache.NewRedisArticleCache(cmdable)
	articleRepository := article2.NewArticleRepository(articleDAO, articleCache, userRepository)
	producer := ioc.InitProducer()
	articleService := service.NewArticleService(articleRepository, producer)
	client := ioc.InitETCD()
	interactiveServiceClient := ioc.InitIntrGRPCClientEtcd(client)
	articleHandle := web.NewArticleHandle(articleService, handler, interactiveServiceClient)
	engine := ioc.InitWebserver(v, userHandle, oAuth2WechatHandler, articleHandle)
	interactiveDAO := dao2.NewGORMInteractiveDAO(db)
	interactiveCache := cache2.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository2.NewCachedInteractiveRepository(interactiveDAO, interactiveCache)
	interactiveReadEventConsumer := ioc.InitInteractiveReadEventConsumer(interactiveRepository)
	v2 := ioc.InitConsumers(interactiveReadEventConsumer)
	rankingRedisCache := cache.NewRankingRedisCache(cmdable)
	rankingLocalCache := cache.NewRankingLocalCache()
	rankingRepository := repository.NewOnlyCachedRankingRepository(rankingRedisCache, rankingLocalCache)
	rankingService := service.NewBatchRankingService(articleService, interactiveServiceClient, rankingRepository)
	rlockClient := ioc.InitRlockClient(cmdable)
	rankingJob := ioc.InitRankingJob(rankingService, rlockClient)
	cron := ioc.InitJobs(rankingJob)
	app := &App{
		web:       engine,
		consumers: v2,
		cron:      cron,
	}
	return app
}

// wire.go:

var rankingSvcSet = wire.NewSet(cache.NewRankingRedisCache, cache.NewRankingLocalCache, repository.NewOnlyCachedRankingRepository, service.NewBatchRankingService)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO, cache2.NewInteractiveRedisCache, repository2.NewCachedInteractiveRepository, service2.NewInteractiveService)
