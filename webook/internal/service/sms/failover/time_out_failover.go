package failover

import (
	"basic-project/webook/internal/service/sms"
	"context"
	"errors"
	"log"
	"sync/atomic"
)

type TimeOutFailoverSMSService struct {
	svcs      []sms.Service
	cnt       uint32
	idx       uint32
	threshold uint32
}

func (t *TimeOutFailoverSMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	cnt := atomic.LoadUint32(&t.cnt)
	idx := atomic.LoadUint32(&t.idx)
	if cnt > t.threshold {
		newIdx := (idx + 1) % uint32(len(t.svcs))
		if atomic.CompareAndSwapUint32(&t.idx, idx, newIdx) {
			atomic.StoreUint32(&t.cnt, 0)
		}
	}
	svc := t.svcs[int(idx)]
	err := svc.Send(ctx, biz, args, numbers...)
	switch {
	case err == nil:
		atomic.StoreUint32(&t.cnt, 0)
	case errors.Is(err, context.DeadlineExceeded):
		atomic.AddUint32(&t.cnt, 1)
	default:
		log.Println("保持不动")
	}
	return nil
}
