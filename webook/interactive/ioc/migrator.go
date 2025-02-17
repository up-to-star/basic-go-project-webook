package ioc

import (
	"github.com/basic-go-project-webook/webook/interactive/repository/dao"
	"github.com/basic-go-project-webook/webook/pkg/ginx"
	"github.com/basic-go-project-webook/webook/pkg/gormx/connpool"
	"github.com/basic-go-project-webook/webook/pkg/migrator/events"
	"github.com/basic-go-project-webook/webook/pkg/migrator/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

func InitGinxServer(src SrcDB, dst DstDB, pool *connpool.DoubleWritePool, producer events.Producer) *ginx.Server {

	engine := gin.Default()
	group := engine.Group("/migrator/interactive")
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "study",
		Subsystem: "webook",
		Name:      "biz_code",
		Help:      "统计业务错误",
	})
	sch := scheduler.NewScheduler[dao.Interactive](src, dst, pool, producer)
	sch.RegisterRoutes(group)
	return &ginx.Server{
		Addr:   viper.GetString("migrator.http.addr"),
		Engine: engine,
	}
}
