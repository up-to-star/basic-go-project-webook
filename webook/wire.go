//go:build wireinject

package main

import (
	"basic-project/webook/internal/repository"
	"basic-project/webook/internal/repository/cache"
	"basic-project/webook/internal/repository/dao"
	"basic-project/webook/internal/service"
	"basic-project/webook/internal/web"
	ijwt "basic-project/webook/internal/web/jwt"
	"basic-project/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		ioc.InitDB, ioc.InitRedis,
		// dao 部分
		dao.NewUserDAO,
		dao.NewArticleDAO,
		// cache 部分
		cache.NewUserCache, cache.NewCodeCache,
		// repository
		repository.NewUserRepository, repository.NewCodeRepository,
		repository.NewArticleRepository,
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
	)
	return gin.Default()
}
