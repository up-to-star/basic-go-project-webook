// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"basic-project/webook/internal/repository"
	article2 "basic-project/webook/internal/repository/article"
	"basic-project/webook/internal/repository/cache"
	"basic-project/webook/internal/repository/dao"
	"basic-project/webook/internal/repository/dao/article"
	"basic-project/webook/internal/service"
	"basic-project/webook/internal/web"
	"basic-project/webook/internal/web/jwt"
	"basic-project/webook/ioc"
	"github.com/gin-gonic/gin"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJwtHandler(cmdable)
	v := ioc.InitGinMiddlewares(cmdable, handler)
	db := ioc.InitDBDefault()
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
	articleRepository := article2.NewArticleRepository(articleDAO)
	articleService := service.NewArticleService(articleRepository)
	articleHandle := web.NewArticleHandle(articleService, handler)
	engine := ioc.InitWebserver(v, userHandle, oAuth2WechatHandler, articleHandle)
	return engine
}
