package ioc

import (
	"github.com/basic-go-project-webook/webook/interactive/repository/dao"
	"github.com/basic-go-project-webook/webook/pkg/gormx"
	"github.com/basic-go-project-webook/webook/pkg/gormx/connpool"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
	"moul.io/zapgorm2"
)

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitSrcDB() SrcDB {
	return initDB("src")
}

func InitDstDB() DstDB {
	return initDB("dst")
}

func InitDoubleWritePool(src SrcDB, dst DstDB) *connpool.DoubleWritePool {
	return connpool.NewDoubleWritePool(src, dst)
}

func InitBizDB(p *connpool.DoubleWritePool) *gorm.DB {
	logger := zapgorm2.New(zap.L())
	logger.LogMode(gormlogger.Info)
	logger.SetAsDefault()
	doubleWrite, err := gorm.Open(mysql.New(mysql.Config{
		Conn: p,
	}), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		panic(err)
	}
	return doubleWrite
}

func initDB(key string) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config
	err := viper.UnmarshalKey("db."+key, &cfg)
	if err != nil {
		panic(err)
	}
	logger := zapgorm2.New(zap.L())
	logger.SetAsDefault()
	logger.LogMode(gormlogger.Info)
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		panic(err)
	}
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook_" + key,
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

	err = db.Use(tracing.NewPlugin(tracing.WithDBName("webook_" + key)))
	if err != nil {
		panic(err)
	}

	pcb := gormx.NewCallbacks(prometheus2.SummaryOpts{
		Namespace: "study_webook",
		Subsystem: "webook",
		Name:      "gorm_db_" + key,
		Help:      "统计 GORM 执行时间",
		ConstLabels: map[string]string{
			"db": "webook",
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
		DBName:          "webook",
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
	err = db.Use(tracing.NewPlugin(tracing.WithDBName("webook")))
	if err != nil {
		panic(err)
	}

	// 监控查询的执行时间
	pcb := gormx.NewCallbacks(prometheus2.SummaryOpts{
		Namespace: "study_webook",
		Subsystem: "webook",
		Name:      "gorm_db",
		Help:      "统计 GORM 执行时间",
		ConstLabels: map[string]string{
			"db": "webook",
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

func InitDBDefault() *gorm.DB {
	logger := zapgorm2.New(zap.L())
	logger.SetAsDefault()
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook_interactive?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}

	return db
}
