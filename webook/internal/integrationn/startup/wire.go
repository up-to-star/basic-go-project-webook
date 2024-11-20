//go:build wireinject

package startup

import (
	"basic-project/webook/internal/repository"
	"basic-project/webook/internal/repository/cache"
	"basic-project/webook/internal/repository/dao"
	"basic-project/webook/internal/service"
	"basic-project/webook/internal/web"
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
		// cache 部分
		cache.NewUserCache, cache.NewCodeCache,
		// repository
		repository.NewUserRepository, repository.NewCodeRepository,
		// service 部分
		ioc.InitSMSService,
		service.NewUserService,
		service.NewCodeService,
		// handler 部分
		web.NewUserHandle,
		ioc.InitGinMiddlewares,
		ioc.InitWebserver,
	)
	return gin.Default()
}
