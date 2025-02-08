package job

import (
	"context"
	"errors"
	"github.com/basic-go-project-webook/webook/internal/domain"
	"github.com/basic-go-project-webook/webook/internal/service"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"time"
)

type Executor interface {
	Name() string
	Execute(ctx context.Context, job domain.Job) error
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, job domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		funcs: make(map[string]func(ctx context.Context, job domain.Job) error),
	}
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, job domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Execute(ctx context.Context, job domain.Job) error {
	fn, ok := l.funcs[job.Name]
	if !ok {
		return errors.New("未注册本地服务方法: " + job.Name)
	}
	return fn(ctx, job)
}

type Scheduler struct {
	svc service.CronJobService
	// 限制并发量
	limiter   *semaphore.Weighted
	dbTimeout time.Duration
	executors map[string]Executor
}

func NewScheduler(svc service.CronJobService) *Scheduler {
	return &Scheduler{
		svc:       svc,
		limiter:   semaphore.NewWeighted(100),
		executors: make(map[string]Executor),
	}
}

func (s *Scheduler) RegisterExecutor(executor Executor) {
	s.executors[executor.Name()] = executor
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		// 放弃调度
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		job, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 有错误，直接进入下一轮抢占
			continue
		}
		executor, ok := s.executors[job.Executor]
		if !ok {
			zap.L().Error("未注册的执行器", zap.String("executor", job.Executor))
			continue
		}
		go func() {
			defer func() {
				s.limiter.Release(1)
				job.CancelFunc()
			}()
			err1 := executor.Execute(ctx, job)
			if err1 != nil {
				zap.L().Error("执行任务执行失败", zap.Int64("jid", job.Id), zap.Error(err1))
				return
			}
			err1 = s.svc.ResetNextTime(ctx, job)
			if err1 != nil {
				zap.L().Error("重置下次执行时间失败", zap.Int64("jid", job.Id), zap.Error(err1))
			}
		}()
	}
}
