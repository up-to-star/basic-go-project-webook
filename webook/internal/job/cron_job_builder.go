package job

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type CronJobBuilder struct {
	vector *prometheus.SummaryVec
}

func NewCronJobBuilder() *CronJobBuilder {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "study",
		Subsystem: "webook_cron_job",
		Name:      "cron_job",
		Help:      "统计 cron job 执行情况",
		Objectives: map[float64]float64{
			0.5:  0.01,
			0.75: 0.01,
			0.9:  0.01,
			0.99: 0.001,
		},
	}, []string{"job", "success"})
	return &CronJobBuilder{
		vector: vector,
	}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()

	return cronJobAdapterFunc(func() {
		start := time.Now()
		zap.L().Info("开始执行 cron job", zap.String("job", name))
		err := job.Run()
		if err != nil {
			zap.L().Error("执行 cron job 失败", zap.String("job", name), zap.Error(err))
		}
		zap.L().Info("执行 cron job 完成", zap.String("job", name))
		duration := time.Since(start)
		b.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).Observe(float64(duration.Milliseconds()))
	})

}

type cronJobAdapterFunc func()

func (c cronJobAdapterFunc) Run() {
	c()
}
