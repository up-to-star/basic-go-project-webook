package service

import (
	"basic-project/webook/internal/repository"
	"basic-project/webook/internal/service/sms"
	"context"
	"fmt"
	"math/rand"
)

var (
	ErrCodeSendTooMany   = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooMany = repository.ErrCodeVerifyTooMany
)

const templateId = "123456"

type CodeService struct {
	repo   *repository.CodeRepository
	smsSvc sms.Service
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.Service) *CodeService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

// Send biz 区别使用的业务
func (svc *CodeService) Send(ctx context.Context, biz string, phone string) error {
	code := svc.generateCode()
	// 放入 redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	err = svc.smsSvc.Send(ctx, templateId, []string{code}, phone)
	//if err != nil {
	//	// 发送失败，redis里面有code, 可以重试
	//}
	return err
}

func (svc *CodeService) Verify(ctx context.Context, biz string, phong string, code string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phong, code)
}

func (svc *CodeService) generateCode() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%6d", num)
}
