package job

import (
	"context"
	"github.com/basic-go-project-webook/webook/internal/service"
	rlock "github.com/gotomicro/redis-lock"
	"go.uber.org/zap"
	"sync"
	"time"
)

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration
	client    *rlock.Client
	key       string
	localLock *sync.Mutex
	lock      *rlock.Lock
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	lock := r.lock
	if lock == nil {
		// 抢分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeout,
			&rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
				// 重试的超时
			}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			zap.L().Warn("获取分布式锁失败", zap.Error(err))
			return nil
		}
		r.lock = lock
		r.localLock.Unlock()
		go func() {
			// 并不是非得一半就续约
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				// 续约失败了
				// 你也没办法中断当下正在调度的热榜计算（如果有）
				r.localLock.Lock()
				r.lock = nil
				//lock.Unlock()
				r.localLock.Unlock()
			}
		}()
	}
	// 拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

func NewRankingJob(svc service.RankingService, client *rlock.Client, timeout time.Duration) *RankingJob {
	return &RankingJob{
		svc:       svc,
		key:       "job:ranking",
		client:    client,
		localLock: &sync.Mutex{},
		timeout:   timeout,
	}
}
