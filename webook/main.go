package main

import (
	"basic-project/webook/internal/pkg/ginx/middlewares/ratelimit"
	"basic-project/webook/internal/repository"
	"basic-project/webook/internal/repository/dao"
	"basic-project/webook/internal/service"
	"basic-project/webook/internal/web"
	"basic-project/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	rds "github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

func main() {
	//db := initDB()
	//u := initUser(db)
	//server := initWebServer()
	//u.RegisterRoutes(server)
	server := gin.Default()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello world")
	})
	_ = server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		//AllowOrigins: []string{"http://localhost:3000"},
		//AllowMethods: []string{"PUT", "PATCH", "POST"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"x-jwt-token"},
		// 是否允许带 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.Contains(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "www.xxx.com")
		},
		MaxAge: 12 * time.Hour,
	}))
	//store := memstore.NewStore([]byte("aIAcX#$8TT}.J+fE}Oa%l6{|-h{oo4)k"), []byte("ll1K9(4lACcUKN5'G};Ug5l*u>.,][_c"))
	store, err := redis.NewStore(16, "tcp", "localhost:6380", "",
		[]byte("aIAcX#$8TT}.J+fE}Oa%l6{|-h{oo4)k"), []byte("ll1K9(4lACcUKN5'G};Ug5l*u>.,][_c"))
	if err != nil {
		panic(err)
	}
	redisClient := rds.NewClient(&rds.Options{
		Addr: "localhost:6380",
	})

	// 基于redis的限流
	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
	server.Use(sessions.Sessions("mysession", store))
	//server.Use(middleware.NewLoginMiddleWareBuilder().
	//	IgnorePath("/users/login").
	//	IgnorePath("/users/signup").Build())
	server.Use(middleware.NewLoginJWTMiddleWareBuilder().
		IgnorePath("/users/login").
		IgnorePath("/users/signup").Build())
	return server
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook?charset=utf8mb4&parseTime=True&loc=Local"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initUser(db *gorm.DB) *web.UserHandle {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandle(svc)
	return u
}
