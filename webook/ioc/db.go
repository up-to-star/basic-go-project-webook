package ioc

import (
	"basic-project/webook/internal/repository/dao"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"
	"moul.io/zapgorm2"
	"time"
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

	// 监控查询的执行时间
	pcb := newCallbacks()
	pcb.registerAll(db)
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func InitDBDefault() *gorm.DB {
	logger := zapgorm2.New(zap.L())
	logger.SetAsDefault()
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{
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

type Callbacks struct {
	vector *prometheus2.SummaryVec
}

func newCallbacks() *Callbacks {
	vector := prometheus2.NewSummaryVec(prometheus2.SummaryOpts{
		Namespace: "study_webook",
		Subsystem: "webook",
		Name:      "gorm_query_time",
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
	}, []string{"type", "table"})
	prometheus2.MustRegister(vector)
	return &Callbacks{
		vector: vector,
	}
}

func (cb *Callbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (cb *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		startTime, ok := val.(time.Time)
		if !ok {
			return
		}
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		cb.vector.WithLabelValues(typ, table).Observe(float64(time.Since(startTime).Milliseconds()))
	}
}

func (cb *Callbacks) registerAll(db *gorm.DB) {
	err := db.Callback().Create().Before("*").Register("prometheus_before_create", cb.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Create().After("*").Register("prometheus_after_create", cb.after("create"))
	if err != nil {
		panic(err)
	}
	err = db.Callback().Query().Before("*").Register("prometheus_before_query", cb.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Query().After("*").Register("prometheus_after_query", cb.after("query"))
	if err != nil {
		panic(err)
	}
	err = db.Callback().Delete().Before("*").Register("prometheus_before_delete", cb.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Delete().After("*").Register("prometheus_after_delete", cb.after("delete"))
	if err != nil {
		panic(err)
	}
	err = db.Callback().Update().Before("*").Register("prometheus_before_update", cb.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Update().After("*").Register("prometheus_after_update", cb.after("update"))
	if err != nil {
		panic(err)
	}
	err = db.Callback().Row().Before("*").Register("prometheus_before_row", cb.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Row().After("*").Register("prometheus_after_row", cb.after("row"))
	if err != nil {
		panic(err)
	}
}
