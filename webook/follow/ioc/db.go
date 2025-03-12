package ioc

import (
	"github.com/basic-go-project-webook/webook/follow/repository/dao"
	"github.com/basic-go-project-webook/webook/pkg/gormx"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
	"moul.io/zapgorm2"
)

func InitDB() *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config
	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(err)
	}
	logger := zapgorm2.New(zap.L())
	logger.SetAsDefault()
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		panic(err)
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook_follow",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"Threads_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}
	err = db.Use(tracing.NewPlugin(tracing.WithDBName("webook_follow")))
	if err != nil {
		panic(err)
	}

	// 监控查询的执行时间
	pcb := gormx.NewCallbacks(prometheus2.SummaryOpts{
		Namespace: "study_webook",
		Subsystem: "webook_follow",
		Name:      "gorm_query_time",
		Help:      "统计 GORM 执行时间",
		ConstLabels: map[string]string{
			"db": "webook_follow",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	pcb.RegisterAll(db)
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
