package ratelimit

import (
	"basic-project/webook/internal/pkg/ratelimit"
	"basic-project/webook/internal/service/sms"
	"context"
	"fmt"
)

var (
	errLimited = fmt.Errorf("短信服务触发限流")
)

type RatelimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

// Send 装饰器
func (s *RatelimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 之前的工作
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误
		// 可以限流，是一个保守策略，下游服务不靠谱的时候
		// 不限流，下游服务很强
		// 包一下这个错误
		return fmt.Errorf("短信服务判断是否限流出现错误: %w", err)
	}
	if limited {
		return errLimited
	}
	err = s.svc.Send(ctx, tplId, args, numbers...)
	// 之后的操作
	return err
}
