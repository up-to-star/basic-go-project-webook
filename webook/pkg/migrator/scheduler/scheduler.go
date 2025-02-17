package scheduler

import (
	"context"
	"fmt"
	"github.com/basic-go-project-webook/webook/pkg/ginx"
	"github.com/basic-go-project-webook/webook/pkg/gormx/connpool"
	"github.com/basic-go-project-webook/webook/pkg/migrator"
	"github.com/basic-go-project-webook/webook/pkg/migrator/events"
	"github.com/basic-go-project-webook/webook/pkg/migrator/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"sync"
	"time"
)

type Scheduler[T migrator.Entity] struct {
	lock       sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DoubleWritePool
	pattern    string
	producer   events.Producer
	cancelFull func()
	cancelIncr func()

	fulls map[string]func()
}

func NewScheduler[T migrator.Entity](src *gorm.DB, dst *gorm.DB, pool *connpool.DoubleWritePool,
	producer events.Producer) *Scheduler[T] {
	return &Scheduler[T]{
		src:      src,
		dst:      dst,
		pool:     pool,
		producer: producer,
		cancelFull: func() {

		},
		cancelIncr: func() {

		},
		pattern: connpool.PatternSrcOnly,
	}
}

func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	server.POST("/src_only", ginx.Wrap(s.SrcOnly))
	server.POST("/dst_only", ginx.Wrap(s.DstOnly))
	server.POST("/src_first", ginx.Wrap(s.SrcFirst))
	server.POST("/dst_first", ginx.Wrap(s.DstFirst))
	server.POST("/full/start", ginx.Wrap(s.StartFullValidation))
	server.POST("/full/stop", ginx.Wrap(s.StopFullValidation))
	server.POST("/incr/stop", ginx.Wrap(s.StopIncrValidation))
	server.POST("/incr/start", s.StartIncrValidation)
}

func (s *Scheduler[T]) SrcOnly(ctx *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	_ = s.pool.UpdatePattern(connpool.PatternSrcOnly)
	return ginx.Result{
		Code: 0,
		Msg:  "OK",
	}, nil
}

func (s *Scheduler[T]) DstOnly(ctx *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	_ = s.pool.UpdatePattern(connpool.PatternDstOnly)
	return ginx.Result{
		Code: 0,
		Msg:  "OK",
	}, nil
}

func (s *Scheduler[T]) SrcFirst(ctx *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcFirst
	_ = s.pool.UpdatePattern(connpool.PatternSrcFirst)
	return ginx.Result{
		Code: 0,
		Msg:  "OK",
	}, nil
}

func (s *Scheduler[T]) DstFirst(ctx *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	_ = s.pool.UpdatePattern(connpool.PatternDstFirst)
	return ginx.Result{
		Code: 0,
		Msg:  "OK",
	}, nil
}

func (s *Scheduler[T]) StartFullValidation(ctx *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{}, err
	}
	var ctx1 context.Context
	ctx1, s.cancelFull = context.WithCancel(context.Background())
	go func() {
		cancel()
		err := v.Validate(ctx1)
		if err != nil {
			fmt.Println("全量校验失败")
			zap.L().Warn("退出全量校验", zap.Error(err))
		}
	}()
	return ginx.Result{
		Code: 0,
		Msg:  "启动全量校验成功",
	}, nil
}

func (s *Scheduler[T]) StopFullValidation(ctx *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	return ginx.Result{
		Code: 0,
		Msg:  "OK",
	}, nil
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		return validator.NewValidator[T](s.src, s.dst, "SRC", 100, s.producer), nil
	case connpool.PatternDstOnly, connpool.PatternDstFirst:
		return validator.NewValidator[T](s.dst, s.src, "DST", 100, s.producer), nil
	default:
		return nil, fmt.Errorf("未知的 pattern： %s", s.pattern)
	}
}

func (s *Scheduler[T]) StopIncrValidation(ctx *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	return ginx.Result{
		Code: 0,
		Msg:  "OK",
	}, nil
}

func (s *Scheduler[T]) StartIncrValidation(ctx *gin.Context) {
	type StartIntrReq struct {
		Utime    int64 `json:"utime"`
		Interval int64 `json:"interval"`
	}
	var req StartIntrReq
	err := ctx.Bind(&req)
	if err != nil {
		zap.L().Error("输入参数错误", zap.Error(err))
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	cancel := s.cancelIncr
	v, err := s.newValidator()
	if err != nil {
		zap.L().Error("创建校验器失败", zap.Error(err))
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	v.Incr().Utime(req.Utime).SleepInterval(time.Duration(req.Interval) * time.Millisecond)
	go func() {
		var ctx1 context.Context
		ctx1, s.cancelIncr = context.WithCancel(context.Background())
		cancel()
		err := v.Validate(ctx1)
		if err != nil {
			fmt.Println("增量校验失败")
			zap.L().Warn("退出增量校验", zap.Error(err))
		}
	}()
	ctx.JSON(http.StatusOK, ginx.Result{
		Code: 0,
		Msg:  "启动增量校验成功",
	})
}
