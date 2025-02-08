package failover

import (
	"context"
	"errors"
	"github.com/basic-go-project-webook/webook/internal/service/sms"
	"log"
	"sync/atomic"
)

type FailoverSMSService struct {
	svcs []sms.Service
	idx  uint64
}

func NewFailoverSMSService(svcs []sms.Service) sms.Service {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

func (f *FailoverSMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	//for _, svc := range f.svcs {
	//	err := svc.Send(ctx, biz, args, numbers...)
	//	if err == nil {
	//		return nil
	//	}
	//	log.Panicln(err)
	//}
	// 轮训
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		err := f.svcs[int(i%length)].Send(ctx, biz, args, numbers...)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return err
		}
		log.Println(err)
	}
	return errors.New("发送失败，全部服务都失败了")
}
