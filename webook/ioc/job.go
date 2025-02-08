package ioc

import (
	"github.com/basic-go-project-webook/webook/internal/job"
	"github.com/basic-go-project-webook/webook/internal/service"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(svc service.RankingService, client *rlock.Client) *job.RankingJob {
	return job.NewRankingJob(svc, client, time.Second*30)
}

func InitJobs(rjob *job.RankingJob) *cron.Cron {
	builder := job.NewCronJobBuilder()
	expr := cron.New(cron.WithSeconds())
	_, err := expr.AddJob("@every 3s", builder.Build(rjob))
	if err != nil {
		panic(err)
	}
	return expr
}
