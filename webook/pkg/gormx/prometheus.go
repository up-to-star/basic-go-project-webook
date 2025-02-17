package gormx

import (
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
	"time"
)

type Callbacks struct {
	vector *prometheus2.SummaryVec
}

func NewCallbacks(ops prometheus2.SummaryOpts) *Callbacks {
	vector := prometheus2.NewSummaryVec(ops, []string{"type", "table"})
	prometheus2.MustRegister(vector)
	return &Callbacks{
		vector: vector,
	}
}

func (cb *Callbacks) Before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (cb *Callbacks) After(typ string) func(db *gorm.DB) {
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

func (cb *Callbacks) RegisterAll(db *gorm.DB) {
	err := db.Callback().Create().Before("*").Register("prometheus_before_create", cb.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Create().After("*").Register("prometheus_after_create", cb.After("create"))
	if err != nil {
		panic(err)
	}
	err = db.Callback().Query().Before("*").Register("prometheus_before_query", cb.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Query().After("*").Register("prometheus_after_query", cb.After("query"))
	if err != nil {
		panic(err)
	}
	err = db.Callback().Delete().Before("*").Register("prometheus_before_delete", cb.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Delete().After("*").Register("prometheus_after_delete", cb.After("delete"))
	if err != nil {
		panic(err)
	}
	err = db.Callback().Update().Before("*").Register("prometheus_before_update", cb.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Update().After("*").Register("prometheus_after_update", cb.After("update"))
	if err != nil {
		panic(err)
	}
	err = db.Callback().Row().Before("*").Register("prometheus_before_row", cb.Before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Row().After("*").Register("prometheus_after_row", cb.After("row"))
	if err != nil {
		panic(err)
	}
}
