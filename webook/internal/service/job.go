package service

import (
	"context"
	"github.com/basic-go-project-webook/webook/internal/domain"
	"github.com/basic-go-project-webook/webook/internal/repository"
	"go.uber.org/zap"
	"time"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, job domain.Job) error
}

type cronJobService struct {
	repo            repository.CronJobRepository
	refreshInterval time.Duration
}

func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()
	j.CancelFunc = func() {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := c.repo.Release(ctx, j.Id)
		if err != nil {
			zap.L().Error("释放 job 失败", zap.Int64("id", j.Id), zap.Error(err))
		}
	}
	return j, err
}

func (c *cronJobService) ResetNextTime(ctx context.Context, job domain.Job) error {
	nextTime := job.NextTime()
	return c.repo.UpdateNextTime(ctx, job.Id, nextTime)
}

func (c *cronJobService) refresh(id int64) {
	// 更新一下更新时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateTime(ctx, id)
	if err != nil {
		zap.L().Error("续约失败", zap.Int64("id", id), zap.Error(err))
	}
}

func NewCronJobService(repo repository.CronJobRepository) CronJobService {
	return &cronJobService{
		repo:            repo,
		refreshInterval: time.Minute,
	}
}
