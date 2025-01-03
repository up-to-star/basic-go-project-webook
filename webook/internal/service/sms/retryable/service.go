package retryable

import (
	"basic-project/webook/internal/service/sms"
	"context"
)

type Service struct {
	svc      sms.Service
	retryCnt int
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	err := s.svc.Send(ctx, biz, args, numbers...)
	if err != nil && s.retryCnt < 10 {
		err = s.svc.Send(ctx, biz, args, numbers...)
		s.retryCnt++
	}
	return err
}
